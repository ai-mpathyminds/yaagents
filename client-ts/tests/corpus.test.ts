/**
 * Corpus conformance tests for WI-1yaa.TSC-3.
 *
 * Replays every fixture in `spec/examples/v0.1/` via a mock fetch and asserts
 * that `parseResponse` maps it to the correct `AgenticResult` variant.
 *
 * - Valid fixtures (13): must produce the expected `result.type`.
 * - Invalid fixtures (18): client must NOT throw (non-strict mode is always
 *   resilient); the result type may differ from what a conformant server would
 *   send, but the client must produce *some* AgenticResult.
 *
 * Authority: spec/examples/INDEX.md + ADR PI1-yaa-0002 §5
 */

import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { afterEach, describe, expect, it, vi } from "vitest";

import { YaAgentsClient } from "../src/index.js";
import type { AgenticResult } from "../src/index.js";

// ---------------------------------------------------------------------------
// Path resolution (ESM-safe)
// ---------------------------------------------------------------------------

const CORPUS_DIR = join(
  dirname(fileURLToPath(import.meta.url)),
  "../../spec/examples/v0.1",
);

// ---------------------------------------------------------------------------
// Fixture → mock-response mapping
// ---------------------------------------------------------------------------

// Normative content-type constants (§4 — do not paraphrase)
const CT = {
  JSON: "application/json",
  CLARIFICATION: "application/vnd.yaagents.clarification+json",
  VALIDATION: "application/vnd.yaagents.validation-error+json",
  APPROVAL: "application/vnd.yaagents.approval-required+json",
  CONFLICT: "application/vnd.yaagents.conflict+json",
  ERROR: "application/vnd.yaagents.error+json",
  OPERATION: "application/vnd.yaagents.operation+json",
} as const;

interface CorpusCase {
  file: string;
  status: number;
  contentType: string;
  expected: AgenticResult["type"];
}

/** Valid fixtures: 13 total per INDEX.md. */
const VALID_CASES: CorpusCase[] = [
  // clarification+json (HTTP 400) — 2 valid
  { file: "clarification-required.valid.json",             status: 400, contentType: CT.CLARIFICATION, expected: "clarification_required" },
  { file: "clarification-required.valid.multi-input.json", status: 400, contentType: CT.CLARIFICATION, expected: "clarification_required" },
  // validation-error+json (HTTP 422) — 2 valid
  { file: "validation-failed.valid.json",                  status: 422, contentType: CT.VALIDATION,     expected: "validation_failed" },
  { file: "validation-failed.valid.multi-error.json",      status: 422, contentType: CT.VALIDATION,     expected: "validation_failed" },
  // approval-required+json (HTTP 412) — 2 valid
  { file: "approval-required.valid.json",                  status: 412, contentType: CT.APPROVAL,       expected: "approval_required" },
  { file: "approval-required.valid.long-token.json",       status: 412, contentType: CT.APPROVAL,       expected: "approval_required" },
  // conflict+json (HTTP 409) — 2 valid
  { file: "conflict.valid.json",                           status: 409, contentType: CT.CONFLICT,        expected: "conflict" },
  { file: "conflict.valid.no-resource-id.json",            status: 409, contentType: CT.CONFLICT,        expected: "conflict" },
  // error+json (HTTP 403/424/500) — 3 valid
  { file: "agentic-error.valid.forbidden.json",            status: 403, contentType: CT.ERROR,           expected: "forbidden" },
  { file: "agentic-error.valid.failed-dependency.json",    status: 424, contentType: CT.ERROR,           expected: "failed_dependency" },
  { file: "agentic-error.valid.error.json",                status: 500, contentType: CT.ERROR,           expected: "error" },
  // operation+json (HTTP 202) — 2 valid
  { file: "operation-accepted.valid.json",                 status: 202, contentType: CT.OPERATION,       expected: "accepted" },
  { file: "operation-accepted.valid.absolute-url.json",    status: 202, contentType: CT.OPERATION,       expected: "accepted" },
];

/**
 * Invalid fixtures: 18 total.  The client routes on Content-Type (not body
 * content), so an invalid body under the correct CT still maps to the same
 * result.type — the body fields may just be missing/empty.
 */
const INVALID_CASES: CorpusCase[] = [
  // clarification+json invalid — 3
  { file: "clarification-required.invalid.missing-trace.json",  status: 400, contentType: CT.CLARIFICATION, expected: "clarification_required" },
  { file: "clarification-required.invalid.wrong-type.json",     status: 400, contentType: CT.CLARIFICATION, expected: "clarification_required" },
  { file: "clarification-required.invalid.empty-inputs.json",   status: 400, contentType: CT.CLARIFICATION, expected: "clarification_required" },
  // validation-error+json invalid — 3
  { file: "validation-failed.invalid.missing-trace.json",       status: 422, contentType: CT.VALIDATION,     expected: "validation_failed" },
  { file: "validation-failed.invalid.wrong-type.json",          status: 422, contentType: CT.VALIDATION,     expected: "validation_failed" },
  { file: "validation-failed.invalid.missing-message.json",     status: 422, contentType: CT.VALIDATION,     expected: "validation_failed" },
  // approval-required+json invalid — 3
  { file: "approval-required.invalid.missing-trace.json",       status: 412, contentType: CT.APPROVAL,       expected: "approval_required" },
  { file: "approval-required.invalid.wrong-type.json",          status: 412, contentType: CT.APPROVAL,       expected: "approval_required" },
  { file: "approval-required.invalid.missing-approval-token.json", status: 412, contentType: CT.APPROVAL,   expected: "approval_required" },
  // conflict+json invalid — 3
  { file: "conflict.invalid.missing-trace.json",                status: 409, contentType: CT.CONFLICT,        expected: "conflict" },
  { file: "conflict.invalid.wrong-type.json",                   status: 409, contentType: CT.CONFLICT,        expected: "conflict" },
  { file: "conflict.invalid.missing-code.json",                 status: 409, contentType: CT.CONFLICT,        expected: "conflict" },
  // error+json invalid — 3
  { file: "agentic-error.invalid.missing-trace.json",           status: 500, contentType: CT.ERROR,           expected: "error" },
  { file: "agentic-error.invalid.wrong-type.json",              status: 400, contentType: CT.ERROR,           expected: "error" },
  { file: "agentic-error.invalid.empty-code.json",              status: 500, contentType: CT.ERROR,           expected: "error" },
  // operation+json invalid — 3
  { file: "operation-accepted.invalid.missing-trace.json",      status: 202, contentType: CT.OPERATION,       expected: "accepted" },
  { file: "operation-accepted.invalid.wrong-type.json",         status: 202, contentType: CT.OPERATION,       expected: "accepted" },
  { file: "operation-accepted.invalid.missing-operation-id.json", status: 202, contentType: CT.OPERATION,    expected: "accepted" },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function loadFixture(file: string): string {
  return readFileSync(join(CORPUS_DIR, file), "utf-8");
}

function stubFetch(status: number, contentType: string, body: string): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(body, {
          status,
          headers: { "Content-Type": contentType },
        }),
      ),
    ),
  );
}

function makeClient(): YaAgentsClient {
  return new YaAgentsClient({
    baseUrl: "http://localhost:8120",
    token: "test-token",
    tenantId: "tenant-corpus",
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

// ---------------------------------------------------------------------------
// Valid fixture suite (13 fixtures)
// ---------------------------------------------------------------------------

describe("corpus valid fixtures — parseResponse maps to correct result.type", () => {
  for (const { file, status, contentType, expected } of VALID_CASES) {
    it(`${file} → result.type === "${expected}"`, async () => {
      const body = loadFixture(file);
      stubFetch(status, contentType, body);
      const result = await makeClient().campaigns.byId("c1").optimizations().create({});
      expect(result.type).toBe(expected);
    });
  }
});

// ---------------------------------------------------------------------------
// Invalid fixture suite (18 fixtures) — client must never throw
// ---------------------------------------------------------------------------

describe("corpus invalid fixtures — non-strict client never throws", () => {
  for (const { file, status, contentType, expected } of INVALID_CASES) {
    it(`${file} → resolves (no throw); result.type === "${expected}"`, async () => {
      const body = loadFixture(file);
      stubFetch(status, contentType, body);
      const result = await makeClient().campaigns.byId("c1").optimizations().create({});
      // Client must resolve, not reject
      expect(result).toBeDefined();
      // Content-type routing is preserved even with invalid bodies
      expect(result.type).toBe(expected);
    });
  }
});

// ---------------------------------------------------------------------------
// Specific field assertions on selected valid fixtures
// ---------------------------------------------------------------------------

describe("corpus field assertions", () => {
  it("clarification-required.valid.json carries requiredInputs[0].name = successMetric", async () => {
    stubFetch(400, CT.CLARIFICATION, loadFixture("clarification-required.valid.json"));
    const result = await makeClient().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("clarification_required");
    if (result.type === "clarification_required") {
      expect(result.requiredInputs[0]?.name).toBe("successMetric");
      expect(result.trace?.correlationId).toBe("corr-001");
    }
  });

  it("operation-accepted.valid.json carries operationId = op-camp-001-20260517-001", async () => {
    stubFetch(202, CT.OPERATION, loadFixture("operation-accepted.valid.json"));
    const result = await makeClient().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("accepted");
    if (result.type === "accepted") {
      expect(result.operationId).toBe("op-camp-001-20260517-001");
    }
  });

  it("agentic-error.valid.forbidden.json carries code = PERMISSION_DENIED", async () => {
    stubFetch(403, CT.ERROR, loadFixture("agentic-error.valid.forbidden.json"));
    const result = await makeClient().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("forbidden");
    if (result.type === "forbidden") {
      expect(result.code).toBe("PERMISSION_DENIED");
    }
  });

  it("validation-failed.valid.json carries errors[0].field = budget", async () => {
    stubFetch(422, CT.VALIDATION, loadFixture("validation-failed.valid.json"));
    const result = await makeClient().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("validation_failed");
    if (result.type === "validation_failed") {
      expect(result.errors[0]?.field).toBe("budget");
    }
  });

  it("agentic-error.valid.failed-dependency.json maps to failed_dependency", async () => {
    stubFetch(424, CT.ERROR, loadFixture("agentic-error.valid.failed-dependency.json"));
    const result = await makeClient().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("failed_dependency");
    if (result.type === "failed_dependency") {
      expect(result.trace?.correlationId).toBe("corr-402");
    }
  });

  it("PROFILE_VERSION matches yaagents.profile in package.json", async () => {
    const { PROFILE_VERSION } = await import("../src/index.js");
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const pkg = JSON.parse(readFileSync(join(dirname(fileURLToPath(import.meta.url)), "../package.json"), "utf-8"));
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    expect(PROFILE_VERSION).toBe((pkg as { yaagents: { profile: string } }).yaagents.profile);
  });
});
