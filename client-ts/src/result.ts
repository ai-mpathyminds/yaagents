// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

/**
 * `AgenticResult<T>` — discriminated union over every YAAgents Profile v0.1 response type.
 *
 * Default client behaviour is **no-throw**: callers switch on `result.type`.
 * Use `client.strict()` for exception-style handling.
 *
 * ## Exhaustive-switch guarantee
 * Every case in the union has a unique string literal `type` discriminant, so an
 * exhaustive `switch` compiles without a `default` branch:
 *
 * ```ts
 * function handle<T>(r: AgenticResult<T>): void {
 *   switch (r.type) {
 *     case "success":               return use(r.resource);
 *     case "created":               return use(r.resource);
 *     case "accepted":              return poll(r.operationId);
 *     case "clarification_required": return ask(r.requiredInputs);
 *     case "validation_failed":     return fix(r.errors);
 *     case "approval_required":     return await(r.message);
 *     case "forbidden":             return deny(r.message);
 *     case "conflict":              return resolve(r.message);
 *     case "failed_dependency":     return retry(r.dependency);
 *     case "error":                 throw new Error(r.message);
 *     // ← no `default` needed; TypeScript confirms exhaustion
 *   }
 * }
 * ```
 */

// ---------------------------------------------------------------------------
// Shared sub-types
// ---------------------------------------------------------------------------

/** Mandatory trace block present in every agentic response body (§4.1). */
export interface AgenticTrace {
  correlationId: string;
  requestId: string;
}

/** A single required input field returned with `clarification_required`. */
export interface RequiredInput {
  name: string;
  location: string;
  type: string;
  required: boolean;
  question: string;
  allowedValues?: readonly string[];
}

/** A single validation error returned with `validation_failed`. */
export interface ValidationErrorDetail {
  field?: string;
  message: string;
  code?: string;
}

// ---------------------------------------------------------------------------
// Individual variant types (one per row in the normative profile table)
// ---------------------------------------------------------------------------

/** HTTP 200 — plain success; `resource` is the domain payload. */
export type SuccessVariant<T> = {
  readonly type: "success";
  readonly resource: T;
  readonly statusCode: 200;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 201 — resource created; `resource` is the created domain payload. */
export type CreatedVariant<T> = {
  readonly type: "created";
  readonly resource: T;
  readonly statusCode: 201;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 202 — async operation accepted; poll `operationId` for outcome. */
export type AcceptedVariant = {
  readonly type: "accepted";
  readonly operationId: string;
  readonly statusCode: 202;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 400 — agent needs more information before it can proceed. */
export type ClarificationRequiredVariant = {
  readonly type: "clarification_required";
  readonly message: string;
  readonly code: string;
  readonly requiredInputs: readonly RequiredInput[];
  readonly statusCode: 400;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 422 — request inputs failed schema / business-rule validation. */
export type ValidationFailedVariant = {
  readonly type: "validation_failed";
  readonly message: string;
  readonly errors: readonly ValidationErrorDetail[];
  readonly statusCode: 422;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 412 — a human-approval step is required before the operation can run. */
export type ApprovalRequiredVariant = {
  readonly type: "approval_required";
  readonly message: string;
  readonly statusCode: 412;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 403 — caller does not have permission for the requested operation. */
export type ForbiddenVariant = {
  readonly type: "forbidden";
  readonly message: string;
  readonly code: string | undefined;
  readonly statusCode: 403;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 409 — the resource is in a conflicting state. */
export type ConflictVariant = {
  readonly type: "conflict";
  readonly message: string;
  readonly statusCode: 409;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 424 — an upstream dependency the agent relies on is unavailable. */
export type FailedDependencyVariant = {
  readonly type: "failed_dependency";
  readonly message: string;
  readonly code: string | undefined;
  readonly dependency: Record<string, unknown>;
  readonly statusCode: 424;
  readonly trace: AgenticTrace | undefined;
};

/** HTTP 5xx — unrecoverable agent or server error. */
export type ServerErrorVariant = {
  readonly type: "error";
  readonly message: string;
  readonly code: string | undefined;
  readonly statusCode: number;
  readonly trace: AgenticTrace | undefined;
};

// ---------------------------------------------------------------------------
// Composite union types
// ---------------------------------------------------------------------------

/**
 * The three success variants.  `strict()` narrows the caller's result to this
 * union — all three can be narrowed further by switching on `result.type`.
 */
export type AgenticSuccessResult<T> =
  | SuccessVariant<T>
  | CreatedVariant<T>
  | AcceptedVariant;

/** The seven error variants. */
export type AgenticErrorResult =
  | ClarificationRequiredVariant
  | ValidationFailedVariant
  | ApprovalRequiredVariant
  | ForbiddenVariant
  | ConflictVariant
  | FailedDependencyVariant
  | ServerErrorVariant;

/**
 * Discriminated union over **all ten** YAAgents Profile v0.1 response types.
 *
 * `T` is the expected domain type of the success payload.
 * Defaults to `Record<string, unknown>` when omitted.
 */
export type AgenticResult<T = Record<string, unknown>> =
  | AgenticSuccessResult<T>
  | AgenticErrorResult;
