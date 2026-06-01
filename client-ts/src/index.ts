// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

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

// ── Result types ──────────────────────────────────────────────────────────────
export type {
  AgenticResult,
  AgenticSuccessResult,
  AgenticErrorResult,
  AgenticTrace,
  RequiredInput,
  ValidationErrorDetail,
  SuccessVariant,
  CreatedVariant,
  AcceptedVariant,
  ClarificationRequiredVariant,
  ValidationFailedVariant,
  ApprovalRequiredVariant,
  ForbiddenVariant,
  ConflictVariant,
  FailedDependencyVariant,
  ServerErrorVariant,
} from "./result.js";

// ── Typed error classes ───────────────────────────────────────────────────────
export {
  AgenticErrorBase,
  ClarificationRequiredError,
  ValidationFailedError,
  ApprovalRequiredError,
  ForbiddenError,
  ConflictError,
  FailedDependencyError,
  AgenticServerError,
} from "./errors.js";

// ── Strict client ─────────────────────────────────────────────────────────────
export {
  StrictClient,
  unwrapStrict,
  StrictCampaignsAccessor,
  StrictCampaignResource,
  StrictOptimizationsResource,
  StrictAssetsResource,
} from "./strict.js";
