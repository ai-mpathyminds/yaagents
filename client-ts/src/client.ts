/**
 * YaAgentsClient — fetch-based HTTP client for the YAAgents Agentic REST Profile v0.1.
 *
 * Uses the global `fetch` API (Node ≥ 18 / all modern browsers).
 * Zero runtime dependencies.
 *
 * Supports YAAgents Profile v0.1.
 */

import { CampaignsAccessor } from "./resources.js";
import { StrictClient } from "./strict.js";
import type { ClientOptions, RequestOptions } from "./types.js";

export { PROFILE_VERSION } from "./version.js";

/**
 * Minimal HTTP client for the YAAgents Agentic REST Profile.
 *
 * @example
 * ```ts
 * import { YaAgentsClient } from "@aimpathyminds/yaagents-client";
 *
 * const client = new YaAgentsClient({
 *   baseUrl: "https://api.example.com",
 *   token: process.env.YAAGENTS_TOKEN!,
 *   tenantId: "tenant-1",
 * });
 *
 * const result = await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
 * ```
 */
export class YaAgentsClient {
  /**
   * Entry point for campaign resource operations.
   *
   * @example `client.campaigns.byId("c1").optimizations().create(body)`
   */
  readonly campaigns: CampaignsAccessor;

  readonly #baseUrl: string;
  readonly #token: string;
  readonly #tenantId: string;

  constructor({ baseUrl, token, tenantId }: ClientOptions) {
    this.#baseUrl = baseUrl.replace(/\/$/, "");
    this.#token = token;
    this.#tenantId = tenantId;
    this.campaigns = new CampaignsAccessor(this);
  }

  // ---------------------------------------------------------------------------
  // strict() — exception-style wrapper
  // ---------------------------------------------------------------------------

  /**
   * Return a `StrictClient` wrapper where every resource method throws a typed
   * `AgenticErrorBase` subclass instead of returning an error result variant.
   *
   * The non-strict client is unchanged; this creates a parallel view.
   *
   * @example
   * ```ts
   * const strict = client.strict();
   * // throws ClarificationRequiredError if the agent needs more input
   * const result = await strict.campaigns.byId("c1").optimizations().create(body);
   * ```
   */
  strict(): StrictClient {
    return new StrictClient(this);
  }

  // ---------------------------------------------------------------------------
  // Internal request helper — used by resource sub-classes; not public API
  // ---------------------------------------------------------------------------

  /**
   * Send an authenticated HTTP request.
   *
   * Injects `Authorization`, `X-Tenant-ID`, and `X-Correlation-ID` on every call.
   * An `X-Correlation-ID` is auto-generated via `crypto.randomUUID()` unless
   * `options.correlationId` is supplied.
   *
   * @internal
   */
  async _request(
    method: string,
    path: string,
    body?: Record<string, unknown>,
    options?: RequestOptions,
  ): Promise<Response> {
    const url = `${this.#baseUrl}${path}`;
    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.#token}`,
      "X-Tenant-ID": this.#tenantId,
      "X-Correlation-ID": options?.correlationId ?? crypto.randomUUID(),
    };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    return fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  }
}
