// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

/**
 * Tests for WI-1yaa.TSC-2 — AgenticResult<T> discriminated union + strict().
 *
 * AC:
 *   - Exhaustive switch over result.type type-checks (no default needed)
 *   - strict() throws typed error matching the vendor type
 *   - non-strict never throws (returns AgenticResult<T> for all responses)
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import {
  AgenticServerError,
  ApprovalRequiredError,
  ClarificationRequiredError,
  ConflictError,
  FailedDependencyError,
  ForbiddenError,
  StrictClient,
  ValidationFailedError,
  YaAgentsClient,
  unwrapStrict,
} from "../src/index.js";
import type {
  AgenticResult,
  AgenticSuccessResult,
} from "../src/index.js";

// ---------------------------------------------------------------------------
// Compile-time exhaustiveness proof
//
// This function has NO `default` branch and returns `string` from every case.
// TypeScript will refuse to compile it if any AgenticResult variant is missing.
// ---------------------------------------------------------------------------

function exhaustiveSwitch<T>(result: AgenticResult<T>): string {
  switch (result.type) {
    case "success":                return "success";
    case "created":                return "created";
    case "accepted":               return "accepted";
    case "clarification_required": return "clarification_required";
    case "validation_failed":      return "validation_failed";
    case "approval_required":      return "approval_required";
    case "forbidden":              return "forbidden";
    case "conflict":               return "conflict";
    case "failed_dependency":      return "failed_dependency";
    case "error":                  return "error";
    // ← no `default` — TypeScript confirms AgenticResult<T> is fully covered
  }
}

// ---------------------------------------------------------------------------
// Mock fetch helpers
// ---------------------------------------------------------------------------

interface MockResponseInit {
  status: number;
  contentType?: string;
  body?: unknown;
}

function mockFetch({ status, contentType, body }: MockResponseInit) {
  return vi.fn().mockImplementation(() =>
    Promise.resolve(
      new Response(
        body !== undefined ? JSON.stringify(body) : null,
        {
          status,
          headers: contentType ? { "Content-Type": contentType } : {},
        },
      ),
    ),
  );
}

function makeClient(fetchMock: ReturnType<typeof vi.fn>) {
  vi.stubGlobal("fetch", fetchMock);
  return new YaAgentsClient({
    baseUrl: "http://localhost:8120",
    token: "t",
    tenantId: "ten",
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

// ---------------------------------------------------------------------------
// Exhaustive switch — runtime verification
// ---------------------------------------------------------------------------

describe("exhaustive switch (compile-time + runtime)", () => {
  it("covers all 10 result types without a default branch", () => {
    // These calls prove exhaustiveSwitch() handles every variant at runtime.
    // The type-checker already proves no `default` is needed at compile time.
    expect(exhaustiveSwitch({ type: "success",  resource: {}, statusCode: 200, trace: undefined })).toBe("success");
    expect(exhaustiveSwitch({ type: "created",  resource: {}, statusCode: 201, trace: undefined })).toBe("created");
    expect(exhaustiveSwitch({ type: "accepted", operationId: "op-1", statusCode: 202, trace: undefined })).toBe("accepted");
    expect(exhaustiveSwitch({ type: "clarification_required", message: "m", code: "C", requiredInputs: [], statusCode: 400, trace: undefined })).toBe("clarification_required");
    expect(exhaustiveSwitch({ type: "validation_failed",      message: "m", errors: [], statusCode: 422, trace: undefined })).toBe("validation_failed");
    expect(exhaustiveSwitch({ type: "approval_required",      message: "m", statusCode: 412, trace: undefined })).toBe("approval_required");
    expect(exhaustiveSwitch({ type: "forbidden",              message: "m", code: undefined, statusCode: 403, trace: undefined })).toBe("forbidden");
    expect(exhaustiveSwitch({ type: "conflict",               message: "m", statusCode: 409, trace: undefined })).toBe("conflict");
    expect(exhaustiveSwitch({ type: "failed_dependency",      message: "m", code: undefined, dependency: {}, statusCode: 424, trace: undefined })).toBe("failed_dependency");
    expect(exhaustiveSwitch({ type: "error",                  message: "m", code: undefined, statusCode: 500, trace: undefined })).toBe("error");
  });
});

// ---------------------------------------------------------------------------
// parseResponse / non-strict (never throws)
// ---------------------------------------------------------------------------

describe("non-strict: parseResponse maps status + content-type → result variant", () => {
  it("200 application/json → success", async () => {
    const client = makeClient(mockFetch({ status: 200, contentType: "application/json", body: { id: "r1" } }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("success");
    if (result.type === "success") expect(result.statusCode).toBe(200);
  });

  it("201 application/json → created", async () => {
    const client = makeClient(mockFetch({ status: 201, contentType: "application/json", body: { id: "r2" } }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("created");
  });

  it("202 operation+json → accepted with operationId", async () => {
    const client = makeClient(mockFetch({
      status: 202,
      contentType: "application/vnd.yaagents.operation+json",
      body: { operationId: "op-42", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("accepted");
    if (result.type === "accepted") expect(result.operationId).toBe("op-42");
  });

  it("400 clarification+json → clarification_required with requiredInputs", async () => {
    const client = makeClient(mockFetch({
      status: 400,
      contentType: "application/vnd.yaagents.clarification+json",
      body: {
        type: "clarification_required",
        code: "CLARIFICATION_REQUIRED",
        message: "Need metric",
        requiredInputs: [{ name: "metric", location: "body", type: "string", required: true, question: "Which metric?" }],
        trace: { correlationId: "c", requestId: "r" },
      },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("clarification_required");
    if (result.type === "clarification_required") {
      expect(result.requiredInputs).toHaveLength(1);
      expect(result.requiredInputs[0]?.name).toBe("metric");
      expect(result.code).toBe("CLARIFICATION_REQUIRED");
    }
  });

  it("422 validation-error+json → validation_failed with errors", async () => {
    const client = makeClient(mockFetch({
      status: 422,
      contentType: "application/vnd.yaagents.validation-error+json",
      body: { message: "Bad input", errors: [{ field: "goal", message: "required" }], trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("validation_failed");
    if (result.type === "validation_failed") {
      expect(result.errors[0]?.field).toBe("goal");
    }
  });

  it("412 approval-required+json → approval_required", async () => {
    const client = makeClient(mockFetch({
      status: 412,
      contentType: "application/vnd.yaagents.approval-required+json",
      body: { message: "Needs approval", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("approval_required");
  });

  it("403 error+json (type=forbidden) → forbidden", async () => {
    const client = makeClient(mockFetch({
      status: 403,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "forbidden", message: "No permission", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("forbidden");
  });

  it("409 conflict+json → conflict", async () => {
    const client = makeClient(mockFetch({
      status: 409,
      contentType: "application/vnd.yaagents.conflict+json",
      body: { message: "Conflict", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("conflict");
  });

  it("424 error+json (type=failed_dependency) → failed_dependency", async () => {
    const client = makeClient(mockFetch({
      status: 424,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "failed_dependency", message: "Upstream down", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("failed_dependency");
  });

  it("500 error+json → error", async () => {
    const client = makeClient(mockFetch({
      status: 500,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "error", message: "Server error", trace: { correlationId: "c", requestId: "r" } },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("error");
  });

  it("non-strict never throws — even on 500", async () => {
    const client = makeClient(mockFetch({
      status: 500,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "error", message: "boom" },
    }));
    await expect(
      client.campaigns.byId("c1").optimizations().create({}),
    ).resolves.toMatchObject({ type: "error" });
  });

  it("trace block is extracted when present", async () => {
    const client = makeClient(mockFetch({
      status: 400,
      contentType: "application/vnd.yaagents.clarification+json",
      body: {
        type: "clarification_required",
        code: "C",
        message: "m",
        requiredInputs: [],
        trace: { correlationId: "corr-1", requestId: "req-2" },
      },
    }));
    const result = await client.campaigns.byId("c1").optimizations().create({});
    expect(result.trace?.correlationId).toBe("corr-1");
    expect(result.trace?.requestId).toBe("req-2");
  });

  it("assets().generate() also returns AgenticResult", async () => {
    const client = makeClient(mockFetch({ status: 200, contentType: "application/json", body: { url: "s3://..." } }));
    const result = await client.campaigns.byId("c1").assets().generate({});
    expect(result.type).toBe("success");
  });
});

// ---------------------------------------------------------------------------
// unwrapStrict
// ---------------------------------------------------------------------------

describe("unwrapStrict", () => {
  it("returns SuccessVariant unchanged", () => {
    const r: AgenticSuccessResult<{ id: string }> = unwrapStrict({
      type: "success", resource: { id: "x" }, statusCode: 200, trace: undefined,
    });
    expect(r.type).toBe("success");
  });

  it("returns CreatedVariant unchanged", () => {
    const r = unwrapStrict({ type: "created", resource: { id: "y" }, statusCode: 201, trace: undefined });
    expect(r.type).toBe("created");
  });

  it("returns AcceptedVariant unchanged", () => {
    const r = unwrapStrict({ type: "accepted", operationId: "op-1", statusCode: 202, trace: undefined });
    expect(r.type).toBe("accepted");
  });

  it("throws ClarificationRequiredError for clarification_required", () => {
    expect(() => unwrapStrict({ type: "clarification_required", message: "m", code: "C", requiredInputs: [], statusCode: 400, trace: undefined }))
      .toThrowError(ClarificationRequiredError);
  });

  it("ClarificationRequiredError carries requiredInputs", () => {
    try {
      unwrapStrict({ type: "clarification_required", message: "m", code: "C", requiredInputs: [{ name: "n", location: "body", type: "string", required: true, question: "q" }], statusCode: 400, trace: undefined });
    } catch (err) {
      expect(err).toBeInstanceOf(ClarificationRequiredError);
      expect((err as ClarificationRequiredError).requiredInputs).toHaveLength(1);
    }
  });

  it("throws ValidationFailedError for validation_failed", () => {
    expect(() => unwrapStrict({ type: "validation_failed", message: "m", errors: [], statusCode: 422, trace: undefined }))
      .toThrowError(ValidationFailedError);
  });

  it("throws ApprovalRequiredError for approval_required", () => {
    expect(() => unwrapStrict({ type: "approval_required", message: "m", statusCode: 412, trace: undefined }))
      .toThrowError(ApprovalRequiredError);
  });

  it("throws ForbiddenError for forbidden", () => {
    expect(() => unwrapStrict({ type: "forbidden", message: "m", code: undefined, statusCode: 403, trace: undefined }))
      .toThrowError(ForbiddenError);
  });

  it("throws ConflictError for conflict", () => {
    expect(() => unwrapStrict({ type: "conflict", message: "m", statusCode: 409, trace: undefined }))
      .toThrowError(ConflictError);
  });

  it("throws FailedDependencyError for failed_dependency", () => {
    expect(() => unwrapStrict({ type: "failed_dependency", message: "m", code: undefined, dependency: {}, statusCode: 424, trace: undefined }))
      .toThrowError(FailedDependencyError);
  });

  it("throws AgenticServerError for error", () => {
    expect(() => unwrapStrict({ type: "error", message: "boom", code: undefined, statusCode: 500, trace: undefined }))
      .toThrowError(AgenticServerError);
  });
});

// ---------------------------------------------------------------------------
// client.strict() end-to-end
// ---------------------------------------------------------------------------

describe("client.strict()", () => {
  it("returns a StrictClient", () => {
    vi.stubGlobal("fetch", vi.fn());
    const client = new YaAgentsClient({ baseUrl: "http://localhost:8120", token: "t", tenantId: "ten" });
    expect(client.strict()).toBeInstanceOf(StrictClient);
  });

  it("strict client returns success result on 200", async () => {
    const client = makeClient(mockFetch({ status: 200, contentType: "application/json", body: { id: "r1" } }));
    const result = await client.strict().campaigns.byId("c1").optimizations().create({});
    expect(result.type).toBe("success");
  });

  it("strict client throws ClarificationRequiredError on 400", async () => {
    const client = makeClient(mockFetch({
      status: 400,
      contentType: "application/vnd.yaagents.clarification+json",
      body: { type: "clarification_required", code: "C", message: "Need metric", requiredInputs: [], trace: { correlationId: "c", requestId: "r" } },
    }));
    await expect(
      client.strict().campaigns.byId("c1").optimizations().create({}),
    ).rejects.toThrow(ClarificationRequiredError);
  });

  it("strict client throws ValidationFailedError on 422", async () => {
    const client = makeClient(mockFetch({
      status: 422,
      contentType: "application/vnd.yaagents.validation-error+json",
      body: { message: "bad", errors: [] },
    }));
    await expect(
      client.strict().campaigns.byId("c1").optimizations().create({}),
    ).rejects.toThrow(ValidationFailedError);
  });

  it("strict client throws ForbiddenError on 403", async () => {
    const client = makeClient(mockFetch({
      status: 403,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "forbidden", message: "denied" },
    }));
    await expect(
      client.strict().campaigns.byId("c1").optimizations().create({}),
    ).rejects.toThrow(ForbiddenError);
  });

  it("strict assets().generate() also throws on error", async () => {
    const client = makeClient(mockFetch({
      status: 424,
      contentType: "application/vnd.yaagents.error+json",
      body: { type: "failed_dependency", message: "upstream down" },
    }));
    await expect(
      client.strict().campaigns.byId("c1").assets().generate({}),
    ).rejects.toThrow(FailedDependencyError);
  });

  it("thrown error has correct statusCode", async () => {
    const client = makeClient(mockFetch({
      status: 422,
      contentType: "application/vnd.yaagents.validation-error+json",
      body: { message: "bad input", errors: [] },
    }));
    let caught: unknown;
    try {
      await client.strict().campaigns.byId("c1").optimizations().create({});
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(ValidationFailedError);
    expect((caught as ValidationFailedError).statusCode).toBe(422);
  });
});
