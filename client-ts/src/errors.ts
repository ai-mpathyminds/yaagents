// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

/**
 * Typed error classes thrown by `client.strict()` when a non-success
 * `AgenticResult` variant is received.
 *
 * Each class mirrors the discriminant of its corresponding `AgenticResult` variant
 * so callers can `instanceof`-check or switch on `.type`:
 *
 * ```ts
 * try {
 *   await strictClient.campaigns.byId("c1").optimizations().create(body);
 * } catch (err) {
 *   if (err instanceof ClarificationRequiredError) {
 *     console.log(err.requiredInputs);
 *   }
 * }
 * ```
 */

import type {
  AgenticTrace,
  ApprovalRequiredVariant,
  ClarificationRequiredVariant,
  ConflictVariant,
  FailedDependencyVariant,
  ForbiddenVariant,
  RequiredInput,
  ServerErrorVariant,
  ValidationErrorDetail,
  ValidationFailedVariant,
} from "./result.js";

// ---------------------------------------------------------------------------
// Base class
// ---------------------------------------------------------------------------

/** Base class for all typed errors thrown in strict mode. */
export abstract class AgenticErrorBase extends Error {
  /** Discriminant matching the `AgenticResult` variant that caused this error. */
  abstract readonly type: string;
  /** HTTP status code of the upstream response. */
  readonly statusCode: number;
  /** Trace block from the response, if present. */
  readonly trace: AgenticTrace | undefined;

  constructor(
    message: string,
    statusCode: number,
    trace: AgenticTrace | undefined,
  ) {
    super(message);
    this.name = this.constructor.name;
    this.statusCode = statusCode;
    this.trace = trace;
  }
}

// ---------------------------------------------------------------------------
// Concrete error classes — one per error variant
// ---------------------------------------------------------------------------

/** Thrown when the agent requires additional inputs before it can proceed (HTTP 400). */
export class ClarificationRequiredError extends AgenticErrorBase {
  override readonly type = "clarification_required" as const;
  readonly code: string;
  readonly requiredInputs: readonly RequiredInput[];

  constructor(variant: ClarificationRequiredVariant) {
    super(variant.message, variant.statusCode, variant.trace);
    this.code = variant.code;
    this.requiredInputs = variant.requiredInputs;
  }
}

/** Thrown when request inputs fail schema or business-rule validation (HTTP 422). */
export class ValidationFailedError extends AgenticErrorBase {
  override readonly type = "validation_failed" as const;
  readonly errors: readonly ValidationErrorDetail[];

  constructor(variant: ValidationFailedVariant) {
    super(variant.message, variant.statusCode, variant.trace);
    this.errors = variant.errors;
  }
}

/** Thrown when a human-approval step is required (HTTP 412). */
export class ApprovalRequiredError extends AgenticErrorBase {
  override readonly type = "approval_required" as const;

  constructor(variant: ApprovalRequiredVariant) {
    super(variant.message, variant.statusCode, variant.trace);
  }
}

/** Thrown when the caller lacks permission for the operation (HTTP 403). */
export class ForbiddenError extends AgenticErrorBase {
  override readonly type = "forbidden" as const;
  readonly code: string | undefined;

  constructor(variant: ForbiddenVariant) {
    super(variant.message, variant.statusCode, variant.trace);
    this.code = variant.code;
  }
}

/** Thrown when the resource is in a conflicting state (HTTP 409). */
export class ConflictError extends AgenticErrorBase {
  override readonly type = "conflict" as const;

  constructor(variant: ConflictVariant) {
    super(variant.message, variant.statusCode, variant.trace);
  }
}

/** Thrown when an upstream dependency the agent relies on is unavailable (HTTP 424). */
export class FailedDependencyError extends AgenticErrorBase {
  override readonly type = "failed_dependency" as const;
  readonly code: string | undefined;
  readonly dependency: Record<string, unknown>;

  constructor(variant: FailedDependencyVariant) {
    super(variant.message, variant.statusCode, variant.trace);
    this.code = variant.code;
    this.dependency = variant.dependency;
  }
}

/** Thrown for unrecoverable agent or server errors (HTTP 5xx). */
export class AgenticServerError extends AgenticErrorBase {
  override readonly type = "error" as const;
  readonly code: string | undefined;

  constructor(variant: ServerErrorVariant) {
    super(variant.message, variant.statusCode, variant.trace);
    this.code = variant.code;
  }
}
