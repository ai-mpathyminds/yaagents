/**
 * @aimpathyminds/yaagents-client
 *
 * TypeScript client for the YAAgents Agentic REST Profile v0.1.
 * Uses the global `fetch` API — zero runtime dependencies.
 *
 * @example
 * ```ts
 * import { YaAgentsClient, PROFILE_VERSION } from "@aimpathyminds/yaagents-client";
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

export { YaAgentsClient, PROFILE_VERSION } from "./client.js";
export type { ClientOptions, RequestOptions } from "./types.js";
export {
  CampaignResource,
  CampaignsAccessor,
  OptimizationsResource,
  AssetsResource,
} from "./resources.js";
