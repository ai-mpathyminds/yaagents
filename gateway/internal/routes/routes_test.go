package routes_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"
)

// writeTempYAML creates a temp file with the given YAML content and returns its path.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "routes-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing temp file: %v", err)
	}
	return f.Name()
}

const validYAML = `
routes:
  - id: list-campaigns
    method: GET
    path: /campaigns
    target: http://campaign-api:8080
  - id: create-optimization
    method: POST
    path: /campaigns/{campaignId}/optimizations
    target: http://campaign-api:8080
    roles:
      - campaign.manager
    tenantRequired: true
    audit: true
`

func TestLoad_ValidRoutes(t *testing.T) {
	f := writeTempYAML(t, validYAML)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(rs))
	}
}

func TestLoad_PlaceholdersParsed(t *testing.T) {
	f := writeTempYAML(t, validYAML)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := rs[1]
	if len(r.PathParams) != 1 || r.PathParams[0] != "campaignId" {
		t.Fatalf("expected PathParams=[campaignId], got %v", r.PathParams)
	}
}

func TestLoad_NoPlaceholders(t *testing.T) {
	f := writeTempYAML(t, validYAML)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := rs[0]
	if len(r.PathParams) != 0 {
		t.Fatalf("expected no PathParams, got %v", r.PathParams)
	}
}

func TestLoad_MultiplePlaceholders(t *testing.T) {
	y := `
routes:
  - id: get-opt
    method: GET
    path: /campaigns/{campaignId}/optimizations/{optId}
    target: http://api:8080
`
	f := writeTempYAML(t, y)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs[0].PathParams) != 2 {
		t.Fatalf("expected 2 path params, got %v", rs[0].PathParams)
	}
	if rs[0].PathParams[0] != "campaignId" || rs[0].PathParams[1] != "optId" {
		t.Fatalf("unexpected PathParams: %v", rs[0].PathParams)
	}
}

func TestLoad_DuplicateID(t *testing.T) {
	y := `
routes:
  - id: dup
    method: GET
    path: /a
    target: http://a:8080
  - id: dup
    method: POST
    path: /b
    target: http://b:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for duplicate id, got nil")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Errorf("error should mention duplication, got: %v", err)
	}
}

func TestLoad_MissingID(t *testing.T) {
	y := `
routes:
  - method: GET
    path: /a
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestLoad_MissingMethod(t *testing.T) {
	y := `
routes:
  - id: test
    path: /a
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for missing method")
	}
}

func TestLoad_InvalidMethod(t *testing.T) {
	y := `
routes:
  - id: test
    method: FETCH
    path: /a
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for invalid method FETCH")
	}
}

func TestLoad_MethodCaseNormalized(t *testing.T) {
	y := `
routes:
  - id: test
    method: post
    path: /a
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rs[0].Method != "POST" {
		t.Errorf("method should be normalized to POST, got %q", rs[0].Method)
	}
}

func TestLoad_MissingPath(t *testing.T) {
	y := `
routes:
  - id: test
    method: GET
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestLoad_MissingTarget(t *testing.T) {
	y := `
routes:
  - id: test
    method: GET
    path: /a
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for missing target")
	}
}

func TestLoad_InvalidTarget(t *testing.T) {
	y := `
routes:
  - id: test
    method: GET
    path: /a
    target: not-a-url
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for invalid target URL")
	}
}

func TestLoad_EmptyPlaceholder(t *testing.T) {
	y := `
routes:
  - id: test
    method: GET
    path: /a/{}
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for empty placeholder {}")
	}
}

func TestLoad_DuplicatePlaceholder(t *testing.T) {
	y := `
routes:
  - id: test
    method: GET
    path: /a/{id}/b/{id}
    target: http://a:8080
`
	f := writeTempYAML(t, y)
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for duplicate placeholder {id}")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := routes.Load("/nonexistent/path/routes.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	f := writeTempYAML(t, "routes: [{{invalid yaml")
	_, err := routes.Load(f)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoad_RolesAndFlags(t *testing.T) {
	f := writeTempYAML(t, validYAML)
	rs, err := routes.Load(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := rs[1]
	if len(r.Roles) != 1 || r.Roles[0] != "campaign.manager" {
		t.Errorf("unexpected roles: %v", r.Roles)
	}
	if !r.TenantRequired {
		t.Error("tenantRequired should be true")
	}
	if !r.Audit {
		t.Error("audit should be true")
	}
}
