/**
 * Public type exports for the YAAgents TypeScript client.
 *
 * Supports YAAgents Profile v0.1.
 */

/** Options accepted by every resource mutation method. */
export interface RequestOptions {
  /**
   * Override the auto-generated `X-Correlation-ID` header.
   * When omitted, the client generates a fresh UUID for each request.
   */
  correlationId?: string;
}

/** Constructor options for {@link YaAgentsClient}. */
export interface ClientOptions {
  /** Root URL of the YAAgents gateway. Trailing slashes are stripped automatically. */
  baseUrl: string;
  /** Bearer token sent as `Authorization: Bearer <token>` on every request. */
  token: string;
  /** Tenant identifier injected as `X-Tenant-ID` on every request. */
  tenantId: string;
}
