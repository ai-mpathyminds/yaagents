/**
 * ts_client_e2e.mjs — TS client live integration check (WI-1yaa.EX-4).
 *
 * Runs the TypeScript client against the live Compose demo gateway and asserts
 * typed agentic handling:
 *   1. Non-strict mode returns a discriminated AgenticResult (type: "created")
 *   2. Non-strict mode returns type: "forbidden" when role is missing (RBAC)
 *   3. Strict mode throws ForbiddenError for the same request
 *   4. POST /campaigns without successMetric → 400 body contains
 *      type: "clarification_required" (verifies gateway propagates vendor body)
 *
 * Usage:
 *   node scripts/ts_client_e2e.mjs <gateway-url> <jwt-token> <jwt-token-no-role>
 *
 * Exit code: 0 = all assertions pass; 1 = failure (error written to stderr).
 *
 * The script imports the already-built TS client dist:
 *   ../../client-ts/dist/index.mjs   (relative to examples/campaign-api/)
 */

import { createHmac } from "node:crypto";
import { fileURLToPath, pathToFileURL } from "node:url";
import path from "node:path";

// ---------------------------------------------------------------------------
// Locate the built TS client dist (repo-relative)
// ---------------------------------------------------------------------------
const __filename = fileURLToPath(import.meta.url);
const __dir = path.dirname(__filename);
// scripts/ → campaign-api/ → examples/ → yaagents/ → client-ts/dist/
const clientDistPath = path.resolve(__dir, "../../../client-ts/dist/index.mjs");
// On Windows, Node ESM requires file:// URLs for absolute paths (drive letters).
const clientDistUrl = pathToFileURL(clientDistPath).href;

let YaAgentsClient, ForbiddenError, ClarificationRequiredError;
try {
  const mod = await import(clientDistUrl);
  YaAgentsClient = mod.YaAgentsClient;
  ForbiddenError = mod.ForbiddenError;
  ClarificationRequiredError = mod.ClarificationRequiredError;
} catch (err) {
  console.error(
    `FAIL: cannot import TS client from ${clientDistUrl}\n  ${err.message}`,
  );
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Args + JWT helpers
// ---------------------------------------------------------------------------
const gatewayUrl = process.argv[2] || "http://localhost:8120";
const DEMO_SECRET = "demo-secret-not-for-production";

function b64url(buf) {
  return buf.toString("base64url");
}

function makeJwt(secret, roles = ["campaign:optimize"], expDelta = 300) {
  const header = b64url(Buffer.from(JSON.stringify({ alg: "HS256", typ: "JWT" })));
  const payload = b64url(
    Buffer.from(
      JSON.stringify({
        sub: "ts-e2e-tester",
        roles,
        exp: Math.floor(Date.now() / 1000) + expDelta,
      }),
    ),
  );
  const signing = `${header}.${payload}`;
  const sig = b64url(
    createHmac("sha256", secret).update(signing).digest(),
  );
  return `${signing}.${sig}`;
}

const tokenFull = makeJwt(DEMO_SECRET, ["campaign:optimize"]);
const tokenNoRole = makeJwt(DEMO_SECRET, []);  // no campaign:optimize

// ---------------------------------------------------------------------------
// Assertion helper
// ---------------------------------------------------------------------------
let passed = 0;
let failed = 0;

function assert(condition, label, detail = "") {
  if (condition) {
    console.log(`  ✓ ${label}`);
    passed++;
  } else {
    console.error(`  ✗ FAIL: ${label}${detail ? `\n      ${detail}` : ""}`);
    failed++;
  }
}

// ---------------------------------------------------------------------------
// Step 0: create a campaign through the gateway (raw fetch — not client)
// ---------------------------------------------------------------------------
console.log("\n[TS client E2E] Creating fixture campaign via gateway…");

let campaignId;
try {
  const resp = await fetch(`${gatewayUrl}/campaigns`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${tokenFull}`,
      "X-Tenant-ID": "ts-e2e-tenant",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name: "TS E2E Campaign",
      budget: 5000,
      targetAudience: "developers",
      successMetric: "ctr",
    }),
  });
  if (resp.status !== 201) {
    console.error(`FAIL: expected 201 creating campaign, got ${resp.status}`);
    process.exit(1);
  }
  const body = await resp.json();
  campaignId = body.campaign?.id;
  if (!campaignId) {
    console.error("FAIL: no campaign.id in response", JSON.stringify(body));
    process.exit(1);
  }
  console.log(`  campaign created: ${campaignId}`);
} catch (err) {
  console.error(`FAIL: network error creating campaign: ${err.message}`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Step 1: clarification_required — POST /campaigns without successMetric
// ---------------------------------------------------------------------------
console.log("\n[TS client E2E] Check 1: clarification_required body from gateway…");

const clarifResp = await fetch(`${gatewayUrl}/campaigns`, {
  method: "POST",
  headers: {
    Authorization: `Bearer ${tokenFull}`,
    "X-Tenant-ID": "ts-e2e-tenant",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({ name: "ClarifTest", budget: 100, targetAudience: "all" }),
});
assert(clarifResp.status === 400, "clarification_required returns HTTP 400");
const clarifCt = clarifResp.headers.get("content-type") || "";
assert(
  clarifCt.includes("application/vnd.yaagents.clarification+json"),
  "clarification content-type is vendor type",
  `got: ${clarifCt}`,
);
const clarifBody = await clarifResp.json();
assert(
  clarifBody.type === "clarification_required",
  "clarification body.type === clarification_required",
  `got: ${clarifBody.type}`,
);
assert(
  Array.isArray(clarifBody.requiredInputs) && clarifBody.requiredInputs.length > 0,
  "clarification body.requiredInputs non-empty",
);
assert(
  typeof clarifBody.trace?.correlationId === "string",
  "clarification body.trace.correlationId present",
);

// ---------------------------------------------------------------------------
// Step 2: non-strict client — forbidden result (missing role)
// ---------------------------------------------------------------------------
console.log("\n[TS client E2E] Check 2: non-strict forbidden discriminated result…");

const clientNoRole = new YaAgentsClient({
  baseUrl: gatewayUrl,
  token: tokenNoRole,
  tenantId: "ts-e2e-tenant",
});

const forbidResult = await clientNoRole.campaigns
  .byId(campaignId)
  .optimizations()
  .create({ objectives: ["ctr"] });

assert(
  forbidResult.type === "forbidden",
  "non-strict optimizations (no role) → result.type === forbidden",
  `got: ${forbidResult.type}`,
);
assert(
  forbidResult.statusCode === 403,
  "forbidden result.statusCode === 403",
  `got: ${forbidResult.statusCode}`,
);

// ---------------------------------------------------------------------------
// Step 3: strict client — throws ForbiddenError
// ---------------------------------------------------------------------------
console.log("\n[TS client E2E] Check 3: strict mode throws ForbiddenError…");

const strictNoRole = clientNoRole.strict();
let caughtForbidden = false;
let caughtForbiddenInstance = false;
try {
  await strictNoRole.campaigns
    .byId(campaignId)
    .optimizations()
    .create({ objectives: ["ctr"] });
} catch (err) {
  caughtForbidden = true;
  caughtForbiddenInstance = err instanceof ForbiddenError;
}
assert(caughtForbidden, "strict mode throws on forbidden response");
assert(caughtForbiddenInstance, "thrown error instanceof ForbiddenError");

// ---------------------------------------------------------------------------
// Step 4: non-strict client — created result (happy path)
// ---------------------------------------------------------------------------
console.log("\n[TS client E2E] Check 4: non-strict created result (happy path)…");

const clientFull = new YaAgentsClient({
  baseUrl: gatewayUrl,
  token: tokenFull,
  tenantId: "ts-e2e-tenant",
});

const createdResult = await clientFull.campaigns
  .byId(campaignId)
  .optimizations()
  .create({ objectives: ["ctr", "cpl"], maxSuggestions: 2 });

assert(
  createdResult.type === "created",
  "non-strict optimizations (full role) → result.type === created",
  `got: ${createdResult.type}`,
);
assert(
  createdResult.statusCode === 201,
  "created result.statusCode === 201",
  `got: ${createdResult.statusCode}`,
);

// ---------------------------------------------------------------------------
// Summary
// ---------------------------------------------------------------------------
console.log(
  `\n[TS client E2E] ${passed + failed} checks: ${passed} passed, ${failed} failed`,
);
if (failed > 0) {
  process.exit(1);
}
console.log("[TS client E2E] PASS");
