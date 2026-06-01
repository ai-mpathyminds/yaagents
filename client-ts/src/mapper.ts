// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

/**
 * Response mapper: raw `fetch` Response → `AgenticResult<T>` discriminated union.
 *
 * Decision tree (first matching rule wins):
 *  1. HTTP 200                              → `success`
 *  2. HTTP 201                              → `created`
 *  3. `application/vnd.yaagents.operation+json`    → `accepted`
 *  4. `application/vnd.yaagents.clarification+json` → `clarification_required`
 *  5. `application/vnd.yaagents.validation-error+json` → `validation_failed`
 *  6. `application/vnd.yaagents.approval-required+json` → `approval_required`
 *  7. `application/vnd.yaagents.conflict+json`      → `conflict`
 *  8. `application/vnd.yaagents.error+json` (type=forbidden)         → `forbidden`
 *  9. `application/vnd.yaagents.error+json` (type=failed_dependency) → `failed_dependency`
 * 10. everything else                        → `error`
 */

import type {
  AgenticResult,
  AgenticTrace,
  RequiredInput,
  ValidationErrorDetail,
} from "./result.js";

// ---------------------------------------------------------------------------
// Vendor content-type constants (normative table §4 — do NOT paraphrase)
// ---------------------------------------------------------------------------

const CT_OPERATION = "application/vnd.yaagents.operation+json";
const CT_CLARIFICATION = "application/vnd.yaagents.clarification+json";
const CT_VALIDATION = "application/vnd.yaagents.validation-error+json";
const CT_APPROVAL = "application/vnd.yaagents.approval-required+json";
const CT_CONFLICT = "application/vnd.yaagents.conflict+json";
const CT_ERROR = "application/vnd.yaagents.error+json";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function ct(response: Response): string {
  return ((response.headers.get("content-type") ?? "").split(";")[0] ?? "").trim();
}

async function jsonBody(response: Response): Promise<Record<string, unknown>> {
  try {
    return (await response.json()) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function extractTrace(body: Record<string, unknown>): AgenticTrace | undefined {
  const t = body["trace"];
  if (
    typeof t === "object" &&
    t !== null &&
    "correlationId" in t &&
    "requestId" in t
  ) {
    const tr = t as Record<string, unknown>;
    return {
      correlationId: String(tr["correlationId"] ?? ""),
      requestId: String(tr["requestId"] ?? ""),
    };
  }
  return undefined;
}

function str(v: unknown, fallback = ""): string {
  return v != null ? String(v) : fallback;
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Parse a raw `fetch` `Response` into an `AgenticResult<T>`.
 *
 * Never throws — all error conditions are returned as result variants.
 * Use `unwrapStrict` (from `strict.ts`) to convert to exception style.
 */
export async function parseResponse<T>(
  response: Response,
): Promise<AgenticResult<T>> {
  const status = response.status;
  const contentType = ct(response);

  // ── Success: plain JSON ──────────────────────────────────────────────────
  if (status === 200) {
    const body = await jsonBody(response);
    return {
      type: "success",
      resource: body as T,
      statusCode: 200,
      trace: extractTrace(body),
    };
  }

  if (status === 201) {
    const body = await jsonBody(response);
    return {
      type: "created",
      resource: body as T,
      statusCode: 201,
      trace: extractTrace(body),
    };
  }

  // ── Accepted (async operation) ───────────────────────────────────────────
  if (contentType === CT_OPERATION) {
    const body = await jsonBody(response);
    return {
      type: "accepted",
      operationId: str(body["operationId"]),
      statusCode: 202,
      trace: extractTrace(body),
    };
  }

  // ── Clarification required ───────────────────────────────────────────────
  if (contentType === CT_CLARIFICATION) {
    const body = await jsonBody(response);
    return {
      type: "clarification_required",
      message: str(body["message"], "Clarification required"),
      code: str(body["code"], "CLARIFICATION_REQUIRED"),
      requiredInputs: (body["requiredInputs"] as RequiredInput[]) ?? [],
      statusCode: 400,
      trace: extractTrace(body),
    };
  }

  // ── Validation failed ────────────────────────────────────────────────────
  if (contentType === CT_VALIDATION) {
    const body = await jsonBody(response);
    return {
      type: "validation_failed",
      message: str(body["message"], "Validation failed"),
      errors: (body["errors"] as ValidationErrorDetail[]) ?? [],
      statusCode: 422,
      trace: extractTrace(body),
    };
  }

  // ── Approval required ────────────────────────────────────────────────────
  if (contentType === CT_APPROVAL) {
    const body = await jsonBody(response);
    return {
      type: "approval_required",
      message: str(body["message"], "Approval required"),
      statusCode: 412,
      trace: extractTrace(body),
    };
  }

  // ── Conflict ─────────────────────────────────────────────────────────────
  if (contentType === CT_CONFLICT) {
    const body = await jsonBody(response);
    return {
      type: "conflict",
      message: str(body["message"], "Conflict"),
      statusCode: 409,
      trace: extractTrace(body),
    };
  }

  // ── Error vendor type (forbidden / failed_dependency / error) ────────────
  if (contentType === CT_ERROR) {
    const body = await jsonBody(response);
    const errorType = str(body["type"], "error");
    const message = str(body["message"], `Agentic error (HTTP ${status})`);
    const code = body["code"] != null ? str(body["code"]) : undefined;

    if (errorType === "forbidden") {
      return {
        type: "forbidden",
        message,
        code,
        statusCode: 403,
        trace: extractTrace(body),
      };
    }

    if (errorType === "failed_dependency") {
      return {
        type: "failed_dependency",
        message,
        code,
        dependency: body,
        statusCode: 424,
        trace: extractTrace(body),
      };
    }

    return {
      type: "error",
      message,
      code,
      statusCode: status,
      trace: extractTrace(body),
    };
  }

  // ── Fallback: unrecognised content-type or plain non-2xx ─────────────────
  let message: string;
  try {
    const body = await jsonBody(response);
    message = str(body["message"], `Unexpected response (HTTP ${status})`);
  } catch {
    message = `Unexpected response (HTTP ${status})`;
  }

  return {
    type: "error",
    message,
    code: undefined,
    statusCode: status,
    trace: undefined,
  };
}
