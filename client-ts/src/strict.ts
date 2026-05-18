/**
 * `strict()` — exception-style wrapper for `YaAgentsClient`.
 *
 * Obtain via `client.strict()`.  Resource methods return `AgenticSuccessResult<T>`
 * on success and throw a typed `AgenticErrorBase` subclass on any error variant.
 *
 * ```ts
 * const strict = client.strict();
 *
 * try {
 *   const result = await strict.campaigns.byId("c1").optimizations().create(body);
 *   // result: AgenticSuccessResult<unknown>
 *   if (result.type === "created") console.log(result.resource);
 * } catch (err) {
 *   if (err instanceof ClarificationRequiredError) {
 *     console.log(err.requiredInputs);
 *   }
 * }
 * ```
 */

import {
  AgenticServerError,
  ApprovalRequiredError,
  ClarificationRequiredError,
  ConflictError,
  FailedDependencyError,
  ForbiddenError,
  ValidationFailedError,
} from "./errors.js";
import type { AgenticResult, AgenticSuccessResult } from "./result.js";
import type { RequestOptions } from "./types.js";
// type-only — breaks the circular reference at runtime
import type { YaAgentsClient } from "./client.js";

// ---------------------------------------------------------------------------
// unwrapStrict
// ---------------------------------------------------------------------------

/**
 * Convert a no-throw `AgenticResult<T>` into an exception-style result.
 *
 * Returns `AgenticSuccessResult<T>` on success; throws the corresponding
 * typed `AgenticErrorBase` subclass for every error variant.
 *
 * The exhaustive switch contains no `default` branch — TypeScript confirms
 * all ten variants are handled.
 */
export function unwrapStrict<T>(
  result: AgenticResult<T>,
): AgenticSuccessResult<T> {
  switch (result.type) {
    // ── Success variants — pass through ──────────────────────────────────
    case "success":
    case "created":
    case "accepted":
      return result;

    // ── Error variants — throw typed subclass ─────────────────────────────
    case "clarification_required":
      throw new ClarificationRequiredError(result);
    case "validation_failed":
      throw new ValidationFailedError(result);
    case "approval_required":
      throw new ApprovalRequiredError(result);
    case "forbidden":
      throw new ForbiddenError(result);
    case "conflict":
      throw new ConflictError(result);
    case "failed_dependency":
      throw new FailedDependencyError(result);
    case "error":
      throw new AgenticServerError(result);
    // ← no `default` needed; TypeScript confirms exhaustion of AgenticResult<T>
  }
}

// ---------------------------------------------------------------------------
// Strict resource sub-classes
// ---------------------------------------------------------------------------

/** Strict variant of `OptimizationsResource` — throws on non-success. */
export class StrictOptimizationsResource {
  readonly #client: YaAgentsClient;
  readonly #campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.#campaignId = campaignId;
  }

  async create<T = Record<string, unknown>>(
    body: Record<string, unknown>,
    options?: RequestOptions,
  ): Promise<AgenticSuccessResult<T>> {
    const result = await this.#client.campaigns
      .byId(this.#campaignId)
      .optimizations()
      .create<T>(body, options);
    return unwrapStrict(result);
  }
}

/** Strict variant of `AssetsResource` — throws on non-success. */
export class StrictAssetsResource {
  readonly #client: YaAgentsClient;
  readonly #campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.#campaignId = campaignId;
  }

  async generate<T = Record<string, unknown>>(
    body: Record<string, unknown>,
    options?: RequestOptions,
  ): Promise<AgenticSuccessResult<T>> {
    const result = await this.#client.campaigns
      .byId(this.#campaignId)
      .assets()
      .generate<T>(body, options);
    return unwrapStrict(result);
  }
}

/** Strict variant of `CampaignResource`. */
export class StrictCampaignResource {
  readonly #client: YaAgentsClient;
  readonly campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.campaignId = campaignId;
  }

  optimizations(): StrictOptimizationsResource {
    return new StrictOptimizationsResource(this.#client, this.campaignId);
  }

  assets(): StrictAssetsResource {
    return new StrictAssetsResource(this.#client, this.campaignId);
  }
}

/** Strict variant of `CampaignsAccessor`. */
export class StrictCampaignsAccessor {
  readonly #client: YaAgentsClient;

  constructor(client: YaAgentsClient) {
    this.#client = client;
  }

  byId(id: string): StrictCampaignResource {
    return new StrictCampaignResource(this.#client, id);
  }
}

// ---------------------------------------------------------------------------
// StrictClient
// ---------------------------------------------------------------------------

/**
 * Exception-style wrapper returned by `YaAgentsClient.strict()`.
 *
 * Mirrors the non-strict client's resource accessor tree but each leaf method
 * throws a typed `AgenticErrorBase` subclass instead of returning an error variant.
 */
export class StrictClient {
  /** Entry point for campaign resource operations (strict / exception-style). */
  readonly campaigns: StrictCampaignsAccessor;

  constructor(inner: YaAgentsClient) {
    this.campaigns = new StrictCampaignsAccessor(inner);
  }
}
