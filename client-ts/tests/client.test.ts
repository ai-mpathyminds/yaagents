// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

/**
 * Tests for WI-1yaa.TSC-1 — YaAgentsClient + fluent resource accessors.
 *
 * AC:
 *   - Authorization / X-Tenant-ID / X-Correlation-ID headers injected on every request
 *   - X-Correlation-ID auto-generated (UUID) and overridable per call
 *   - Resource accessors build the correct method + path + body
 *   - PROFILE_VERSION exported as "v0.2"
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  AssetsResource,
  CampaignResource,
  CampaignsAccessor,
  OptimizationsResource,
  PROFILE_VERSION,
  YaAgentsClient,
} from "../src/index.js";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Create a mock fetch that records calls and returns a fresh 200 JSON response each time. */
function makeMockFetch(
  status = 200,
  body: unknown = { ok: true },
): ReturnType<typeof vi.fn> {
  // Use mockImplementation (not mockResolvedValue) so each call gets a NEW
  // Response object — a consumed Response body cannot be read a second time.
  return vi.fn().mockImplementation(() =>
    Promise.resolve(
      new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
    ),
  );
}

/** Extract RequestInit from a mock fetch call. */
function getInit(mockFn: ReturnType<typeof vi.fn>, idx = 0): RequestInit {
  return (mockFn.mock.calls[idx] as [string, RequestInit])[1];
}

/** Extract the request URL from a mock fetch call. */
function getUrl(mockFn: ReturnType<typeof vi.fn>, idx = 0): string {
  return (mockFn.mock.calls[idx] as [string])[0];
}

/** Extract headers record from a mock fetch call. */
function getHeaders(
  mockFn: ReturnType<typeof vi.fn>,
  idx = 0,
): Record<string, string> {
  return getInit(mockFn, idx).headers as Record<string, string>;
}

/** Create a default test client with the given mock fetch already stubbed. */
function makeClient(
  mockFn: ReturnType<typeof vi.fn>,
  overrides: Partial<{ baseUrl: string; token: string; tenantId: string }> = {},
): YaAgentsClient {
  vi.stubGlobal("fetch", mockFn);
  return new YaAgentsClient({
    baseUrl: "http://localhost:8120",
    token: "test-token",
    tenantId: "tenant-1",
    ...overrides,
  });
}

// ---------------------------------------------------------------------------
// Setup / teardown
// ---------------------------------------------------------------------------

afterEach(() => {
  vi.unstubAllGlobals();
});

// ---------------------------------------------------------------------------
// Header injection
// ---------------------------------------------------------------------------

describe("header injection", () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = makeMockFetch();
  });

  it("injects Authorization: Bearer <token>", async () => {
    const client = makeClient(mockFetch, { token: "my-secret-token" });
    await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
    expect(getHeaders(mockFetch)["Authorization"]).toBe(
      "Bearer my-secret-token",
    );
  });

  it("injects X-Tenant-ID", async () => {
    const client = makeClient(mockFetch, { tenantId: "tenant-99" });
    await client.campaigns.byId("c1").assets().generate({ style: "bold" });
    expect(getHeaders(mockFetch)["X-Tenant-ID"]).toBe("tenant-99");
  });

  it("injects Content-Type: application/json when body is provided", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
    expect(getHeaders(mockFetch)["Content-Type"]).toBe("application/json");
  });

  it("all three default headers present on assets:generate", async () => {
    const client = makeClient(mockFetch, {
      token: "bearer-tok",
      tenantId: "t-42",
    });
    await client.campaigns.byId("camp").assets().generate({ format: "png" });
    const h = getHeaders(mockFetch);
    expect(h["Authorization"]).toBe("Bearer bearer-tok");
    expect(h["X-Tenant-ID"]).toBe("t-42");
    expect(h["X-Correlation-ID"]).toBeTruthy();
  });
});

// ---------------------------------------------------------------------------
// Correlation ID
// ---------------------------------------------------------------------------

describe("X-Correlation-ID", () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = makeMockFetch();
  });

  it("auto-generated as UUID v4", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("c1").optimizations().create({});
    const corr = getHeaders(mockFetch)["X-Correlation-ID"];
    expect(corr).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    );
  });

  it("unique per request", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("c1").optimizations().create({});
    await client.campaigns.byId("c1").optimizations().create({});
    const id0 = getHeaders(mockFetch, 0)["X-Correlation-ID"];
    const id1 = getHeaders(mockFetch, 1)["X-Correlation-ID"];
    expect(id0).not.toBe(id1);
  });

  it("overridable on optimizations.create", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns
      .byId("c1")
      .optimizations()
      .create({}, { correlationId: "my-trace-001" });
    expect(getHeaders(mockFetch)["X-Correlation-ID"]).toBe("my-trace-001");
  });

  it("overridable on assets.generate", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns
      .byId("c2")
      .assets()
      .generate({}, { correlationId: "my-trace-002" });
    expect(getHeaders(mockFetch)["X-Correlation-ID"]).toBe("my-trace-002");
  });
});

// ---------------------------------------------------------------------------
// URL routing
// ---------------------------------------------------------------------------

describe("URL routing", () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = makeMockFetch();
  });

  it("POST /campaigns/{id}/optimizations", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("camp-001").optimizations().create({});
    expect(getUrl(mockFetch)).toBe(
      "http://localhost:8120/campaigns/camp-001/optimizations",
    );
    expect(getInit(mockFetch).method).toBe("POST");
  });

  it("POST /campaigns/{id}/assets:generate", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("camp-002").assets().generate({});
    expect(getUrl(mockFetch)).toBe(
      "http://localhost:8120/campaigns/camp-002/assets:generate",
    );
    expect(getInit(mockFetch).method).toBe("POST");
  });

  it("strips trailing slash from baseUrl", async () => {
    const client = makeClient(mockFetch, { baseUrl: "http://localhost:8120/" });
    await client.campaigns.byId("c6").optimizations().create({});
    expect(getUrl(mockFetch)).toBe(
      "http://localhost:8120/campaigns/c6/optimizations",
    );
  });

  it("two campaign IDs route independently", async () => {
    const client = makeClient(mockFetch);
    await client.campaigns.byId("alpha").optimizations().create({});
    await client.campaigns.byId("beta").optimizations().create({});
    expect(getUrl(mockFetch, 0)).toBe(
      "http://localhost:8120/campaigns/alpha/optimizations",
    );
    expect(getUrl(mockFetch, 1)).toBe(
      "http://localhost:8120/campaigns/beta/optimizations",
    );
  });

  it("serialises body as JSON", async () => {
    const client = makeClient(mockFetch);
    const body = { goal: "conversion_rate", budget: 500 };
    await client.campaigns.byId("c3").optimizations().create(body);
    expect(getInit(mockFetch).body).toBe(JSON.stringify(body));
  });
});

// ---------------------------------------------------------------------------
// Resource type assertions
// ---------------------------------------------------------------------------

describe("resource types", () => {
  it("campaigns is a CampaignsAccessor", () => {
    vi.stubGlobal("fetch", makeMockFetch());
    const client = new YaAgentsClient({
      baseUrl: "http://localhost:8120",
      token: "t",
      tenantId: "ten",
    });
    expect(client.campaigns).toBeInstanceOf(CampaignsAccessor);
  });

  it("byId returns CampaignResource", () => {
    vi.stubGlobal("fetch", makeMockFetch());
    const client = new YaAgentsClient({
      baseUrl: "http://localhost:8120",
      token: "t",
      tenantId: "ten",
    });
    expect(client.campaigns.byId("c1")).toBeInstanceOf(CampaignResource);
  });

  it("optimizations() returns OptimizationsResource", () => {
    vi.stubGlobal("fetch", makeMockFetch());
    const client = new YaAgentsClient({
      baseUrl: "http://localhost:8120",
      token: "t",
      tenantId: "ten",
    });
    expect(
      client.campaigns.byId("c1").optimizations(),
    ).toBeInstanceOf(OptimizationsResource);
  });

  it("assets() returns AssetsResource", () => {
    vi.stubGlobal("fetch", makeMockFetch());
    const client = new YaAgentsClient({
      baseUrl: "http://localhost:8120",
      token: "t",
      tenantId: "ten",
    });
    expect(client.campaigns.byId("c1").assets()).toBeInstanceOf(AssetsResource);
  });

  it("campaignId is preserved on CampaignResource", () => {
    vi.stubGlobal("fetch", makeMockFetch());
    const client = new YaAgentsClient({
      baseUrl: "http://localhost:8120",
      token: "t",
      tenantId: "ten",
    });
    const cr = client.campaigns.byId("my-campaign");
    expect(cr.campaignId).toBe("my-campaign");
  });
});

// ---------------------------------------------------------------------------
// Package metadata
// ---------------------------------------------------------------------------

describe("package metadata", () => {
  it("PROFILE_VERSION is 'v0.2'", () => {
    expect(PROFILE_VERSION).toBe("v0.2");
  });
});
