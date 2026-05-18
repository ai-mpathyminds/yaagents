/**
 * Fluent resource accessors for the YAAgents Agentic REST Profile v0.1.
 *
 * Usage:
 * ```ts
 * const result = await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
 * const asset  = await client.campaigns.byId("c1").assets().generate({ format: "banner" });
 * ```
 */

import type { RequestOptions } from "./types.js";
// type-only import breaks the client ↔ resources circular reference at runtime
import type { YaAgentsClient } from "./client.js";

// ---------------------------------------------------------------------------
// OptimizationsResource
// ---------------------------------------------------------------------------

/** Sub-resource for campaign optimizations. */
export class OptimizationsResource {
  readonly #client: YaAgentsClient;
  readonly #campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.#campaignId = campaignId;
  }

  /**
   * POST /campaigns/{id}/optimizations — create a new optimization run.
   *
   * @param body - Request payload (framework-specific; see your gateway's OpenAPI spec).
   * @param options - Per-call options such as a custom `correlationId`.
   * @returns Parsed JSON response body.
   */
  async create(
    body: Record<string, unknown>,
    options?: RequestOptions,
  ): Promise<unknown> {
    const response = await this.#client._request(
      "POST",
      `/campaigns/${this.#campaignId}/optimizations`,
      body,
      options,
    );
    return response.json() as Promise<unknown>;
  }
}

// ---------------------------------------------------------------------------
// AssetsResource
// ---------------------------------------------------------------------------

/** Sub-resource for campaign asset generation. */
export class AssetsResource {
  readonly #client: YaAgentsClient;
  readonly #campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.#campaignId = campaignId;
  }

  /**
   * POST /campaigns/{id}/assets:generate — trigger asset generation.
   *
   * @param body - Request payload.
   * @param options - Per-call options such as a custom `correlationId`.
   * @returns Parsed JSON response body.
   */
  async generate(
    body: Record<string, unknown>,
    options?: RequestOptions,
  ): Promise<unknown> {
    const response = await this.#client._request(
      "POST",
      `/campaigns/${this.#campaignId}/assets:generate`,
      body,
      options,
    );
    return response.json() as Promise<unknown>;
  }
}

// ---------------------------------------------------------------------------
// CampaignResource
// ---------------------------------------------------------------------------

/** Fluent accessor for a single campaign and its sub-resources. */
export class CampaignResource {
  readonly #client: YaAgentsClient;
  /** Campaign identifier used in every sub-resource path. */
  readonly campaignId: string;

  constructor(client: YaAgentsClient, campaignId: string) {
    this.#client = client;
    this.campaignId = campaignId;
  }

  /** Return an {@link OptimizationsResource} scoped to this campaign. */
  optimizations(): OptimizationsResource {
    return new OptimizationsResource(this.#client, this.campaignId);
  }

  /** Return an {@link AssetsResource} scoped to this campaign. */
  assets(): AssetsResource {
    return new AssetsResource(this.#client, this.campaignId);
  }
}

// ---------------------------------------------------------------------------
// CampaignsAccessor
// ---------------------------------------------------------------------------

/**
 * Entry point for the campaigns resource collection.
 * Obtained via `client.campaigns`.
 */
export class CampaignsAccessor {
  readonly #client: YaAgentsClient;

  constructor(client: YaAgentsClient) {
    this.#client = client;
  }

  /**
   * Scope subsequent calls to a specific campaign.
   *
   * @param id - Campaign identifier.
   * @returns A {@link CampaignResource} fluent accessor.
   *
   * @example
   * ```ts
   * await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
   * ```
   */
  byId(id: string): CampaignResource {
    return new CampaignResource(this.#client, id);
  }
}
