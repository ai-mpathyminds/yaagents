package yaagentsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ── Resource accessor chain ────────────────────────────────────────────────
//
// Each resource type is a small struct holding *Client + parent IDs only.
// The chain is: client.Campaigns() → .ByID(id) → .Optimizations()/.Assets()
// → terminal methods (Create / Get / Generate).
//
// PRD §5.9 resource symbols covered here:
//   client.Campaigns() CampaignsResource
//   CampaignsResource.ByID(id) CampaignResource
//   CampaignResource.Optimizations() OptimizationsResource
//   CampaignResource.Assets() AssetsResource
//   OptimizationsResource.Create(ctx, body) (*AgenticResult, error)
//   OptimizationsResource.Get(ctx, id) (*AgenticResult, error)
//   AssetsResource.Generate(ctx, body) (*AgenticResult, error)

// ── CampaignsResource ──────────────────────────────────────────────────────

// CampaignsResource is the entry point for campaign-scoped operations.
type CampaignsResource struct{ client *Client }

// Campaigns returns the CampaignsResource root accessor.
func (c *Client) Campaigns() CampaignsResource {
	return CampaignsResource{client: c}
}

// ByID returns a CampaignResource scoped to the given campaign ID.
func (r CampaignsResource) ByID(campaignID string) CampaignResource {
	return CampaignResource{client: r.client, campaignID: campaignID}
}

// ── CampaignResource ───────────────────────────────────────────────────────

// CampaignResource is scoped to a specific campaign.
type CampaignResource struct {
	client     *Client
	campaignID string
}

// Optimizations returns the OptimizationsResource for this campaign.
func (r CampaignResource) Optimizations() OptimizationsResource {
	return OptimizationsResource{client: r.client, campaignID: r.campaignID}
}

// Assets returns the AssetsResource for this campaign.
func (r CampaignResource) Assets() AssetsResource {
	return AssetsResource{client: r.client, campaignID: r.campaignID}
}

// ── OptimizationsResource ──────────────────────────────────────────────────

// OptimizationsResource exposes Create and Get for campaign optimizations.
type OptimizationsResource struct {
	client     *Client
	campaignID string
}

// Create posts a new optimization request for the campaign.
//
// Endpoint: POST /campaigns/{campaignID}/optimizations
func (r OptimizationsResource) Create(ctx context.Context, body any) (*AgenticResult, error) {
	b, err := marshalBody(body)
	if err != nil {
		return nil, err
	}
	return r.client.do(ctx, http.MethodPost,
		fmt.Sprintf("/campaigns/%s/optimizations", r.campaignID), b)
}

// Get retrieves an optimization result by ID.
//
// Endpoint: GET /campaigns/{campaignID}/optimizations/{optimizationID}
func (r OptimizationsResource) Get(ctx context.Context, optimizationID string) (*AgenticResult, error) {
	return r.client.do(ctx, http.MethodGet,
		fmt.Sprintf("/campaigns/%s/optimizations/%s", r.campaignID, optimizationID), nil)
}

// ── AssetsResource ─────────────────────────────────────────────────────────

// AssetsResource exposes Generate for campaign assets.
type AssetsResource struct {
	client     *Client
	campaignID string
}

// Generate posts an asset-generation request for the campaign.
//
// Endpoint: POST /campaigns/{campaignID}/assets:generate
func (r AssetsResource) Generate(ctx context.Context, body any) (*AgenticResult, error) {
	b, err := marshalBody(body)
	if err != nil {
		return nil, err
	}
	return r.client.do(ctx, http.MethodPost,
		fmt.Sprintf("/campaigns/%s/assets:generate", r.campaignID), b)
}

// ── helpers ────────────────────────────────────────────────────────────────

// marshalBody serialises v to JSON and wraps it in a *bytes.Reader.
// Returns a wrapped error on failure; callers pass it directly to do().
func marshalBody(v any) (*bytes.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("yaagentsclient: marshal body: %w", err)
	}
	return bytes.NewReader(data), nil
}
