package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/auth"
)

// ── HS256 ─────────────────────────────────────────────────────────────────────

func TestHS256Validator_ValidToken(t *testing.T) {
	v := auth.NewHS256Validator("test-secret")
	tok := makeHS256Token(t, "test-secret", "user-1", futureExp(), []string{"admin"})

	claims, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Errorf("Subject: got %q, want user-1", claims.Subject)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("Roles: got %v, want [admin]", claims.Roles)
	}
}

func TestHS256Validator_ExpiredToken(t *testing.T) {
	v := auth.NewHS256Validator("test-secret")
	tok := makeHS256Token(t, "test-secret", "u", pastExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestHS256Validator_TamperedSignature(t *testing.T) {
	v := auth.NewHS256Validator("test-secret")
	// Signed with a different secret — signature mismatch.
	tok := makeHS256Token(t, "wrong-secret", "u", futureExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for tampered signature")
	}
}

func TestHS256Validator_WrongAlgorithm(t *testing.T) {
	v := auth.NewHS256Validator("test-secret")
	// Feed an RS256-signed token to the HS256 validator.
	priv, _ := rsaKeyPair(t)
	tok := makeRS256Token(t, priv, "kid-1", "u", futureExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for wrong algorithm")
	}
}

func TestHS256Validator_RolesInClaims(t *testing.T) {
	v := auth.NewHS256Validator("s")
	tok := makeHS256Token(t, "s", "u", futureExp(), []string{"r1", "r2"})

	claims, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(claims.Roles) != 2 {
		t.Errorf("Roles: got %v, want [r1 r2]", claims.Roles)
	}
}

// ── JWKS / RS256 ──────────────────────────────────────────────────────────────

func TestJWKSValidator_ValidToken(t *testing.T) {
	priv, pub := rsaKeyPair(t)
	kid := "key-1"
	srv := httptest.NewServer(jwksServer(pub, kid))
	defer srv.Close()

	v := auth.NewJWKSValidator(srv.URL)
	tok := makeRS256Token(t, priv, kid, "user-2", futureExp(), []string{"campaign.manager"})

	claims, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "user-2" {
		t.Errorf("Subject: got %q, want user-2", claims.Subject)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "campaign.manager" {
		t.Errorf("Roles: got %v, want [campaign.manager]", claims.Roles)
	}
}

func TestJWKSValidator_ExpiredToken(t *testing.T) {
	priv, pub := rsaKeyPair(t)
	kid := "key-1"
	srv := httptest.NewServer(jwksServer(pub, kid))
	defer srv.Close()

	v := auth.NewJWKSValidator(srv.URL)
	tok := makeRS256Token(t, priv, kid, "u", pastExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWKSValidator_UnknownKid(t *testing.T) {
	_, pub := rsaKeyPair(t)
	srv := httptest.NewServer(jwksServer(pub, "key-1"))
	defer srv.Close()

	// Token signed with a different key not in the JWKS.
	priv2, _ := rsaKeyPair(t)
	tok := makeRS256Token(t, priv2, "unknown-kid", "u", futureExp(), nil)

	v := auth.NewJWKSValidator(srv.URL)
	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for unknown kid")
	}
}

func TestJWKSValidator_WrongAlgorithm(t *testing.T) {
	_, pub := rsaKeyPair(t)
	srv := httptest.NewServer(jwksServer(pub, "k1"))
	defer srv.Close()

	v := auth.NewJWKSValidator(srv.URL)
	// Feed an HS256 token to the JWKS (RS256) validator.
	tok := makeHS256Token(t, "secret", "u", futureExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for wrong algorithm")
	}
}

func TestJWKSValidator_CacheHit(t *testing.T) {
	priv, pub := rsaKeyPair(t)
	kid := "key-1"
	fetchCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		jwksServer(pub, kid).ServeHTTP(w, r)
	}))
	defer srv.Close()

	v := auth.NewJWKSValidator(srv.URL)
	tok := makeRS256Token(t, priv, kid, "u", futureExp(), nil)

	if _, err := v.Validate(tok); err != nil {
		t.Fatalf("first validate: %v", err)
	}
	if _, err := v.Validate(tok); err != nil {
		t.Fatalf("second validate: %v", err)
	}
	// Second call must use the cached keys — only one HTTP fetch.
	if fetchCount != 1 {
		t.Errorf("expected 1 JWKS fetch, got %d", fetchCount)
	}
}

func TestJWKSValidator_ServerUnreachable(t *testing.T) {
	v := auth.NewJWKSValidator("http://127.0.0.1:0") // nothing listening

	priv, _ := rsaKeyPair(t)
	tok := makeRS256Token(t, priv, "k1", "u", futureExp(), nil)

	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error when JWKS server is unreachable")
	}
}

func TestJWKSValidator_EmptyKeySet(t *testing.T) {
	// JWKS returns valid JSON but an empty keys array.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"keys":[]}`)
	}))
	defer srv.Close()

	priv, _ := rsaKeyPair(t)
	tok := makeRS256Token(t, priv, "k1", "u", futureExp(), nil)

	v := auth.NewJWKSValidator(srv.URL)
	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error: no key found for kid")
	}
}

func TestJWKSValidator_InvalidBase64InModulus(t *testing.T) {
	// JWKS key has invalid base64url in the 'n' field.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"keys":[{"kty":"RSA","kid":"k1","n":"!!!","e":"AQAB"}]}`)
	}))
	defer srv.Close()

	priv, _ := rsaKeyPair(t)
	tok := makeRS256Token(t, priv, "k1", "u", futureExp(), nil)

	v := auth.NewJWKSValidator(srv.URL)
	_, err := v.Validate(tok)
	if err == nil {
		t.Fatal("expected error for invalid base64 in JWKS modulus")
	}
}

// ── Middleware ─────────────────────────────────────────────────────────────────

func TestMiddleware_MissingAuthHeader(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	req := httptest.NewRequest("GET", "/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized)
	assertVendorErrorCT(t, rr)
}

func TestMiddleware_WrongAuthScheme(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized)
	assertVendorErrorCT(t, rr)
}

func TestMiddleware_InvalidToken(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized)
	assertVendorErrorCT(t, rr)
}

func TestMiddleware_ValidToken_CallsNext(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	tok := makeHS256Token(t, "s", "u", futureExp(), nil)
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK)
}

func TestMiddleware_ClaimsStoredInContext(t *testing.T) {
	var gotClaims auth.Claims
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := r.Context().Value(auth.ClaimsKey).(auth.Claims); ok {
			gotClaims = c
		}
		w.WriteHeader(http.StatusOK)
	})
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(inner)
	tok := makeHS256Token(t, "s", "user-ctx", futureExp(), []string{"role1"})
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if gotClaims.Subject != "user-ctx" {
		t.Errorf("context Subject: got %q, want user-ctx", gotClaims.Subject)
	}
}

func TestMiddleware_TracePopulated_Generated(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	req := httptest.NewRequest("GET", "/", nil) // no corr-id header
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := decodeErrorBody(t, rr)
	if body["trace"] == nil {
		t.Fatal("trace missing from error body")
	}
	tr := body["trace"].(map[string]interface{})
	if corrID, _ := tr["correlationId"].(string); corrID == "" {
		t.Error("correlationId should be generated (non-empty)")
	}
	if reqID, _ := tr["requestId"].(string); reqID == "" {
		t.Error("requestId should be generated (non-empty)")
	}
}

func TestMiddleware_TracePopulated_Propagated(t *testing.T) {
	h := auth.Middleware(auth.NewHS256Validator("s"), noopLog())(okHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-xyz")
	req.Header.Set("X-Request-ID", "req-xyz")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := decodeErrorBody(t, rr)
	tr := body["trace"].(map[string]interface{})
	if tr["correlationId"] != "corr-xyz" {
		t.Errorf("correlationId: got %v, want corr-xyz", tr["correlationId"])
	}
	if tr["requestId"] != "req-xyz" {
		t.Errorf("requestId: got %v, want req-xyz", tr["requestId"])
	}
}

// ── NewValidator ──────────────────────────────────────────────────────────────

func TestNewValidator_NeitherSet_Error(t *testing.T) {
	t.Setenv("GATEWAY_JWT_SECRET", "")
	t.Setenv("GATEWAY_JWT_JWKS_URL", "")

	_, err := auth.NewValidator(noopLog())
	if err == nil {
		t.Fatal("expected error when neither env var is set")
	}
}

func TestNewValidator_OnlySecret_HS256(t *testing.T) {
	t.Setenv("GATEWAY_JWT_SECRET", "my-secret")
	t.Setenv("GATEWAY_JWT_JWKS_URL", "")

	v, err := auth.NewValidator(noopLog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it validates HS256 tokens correctly.
	tok := makeHS256Token(t, "my-secret", "u", futureExp(), nil)
	if _, err := v.Validate(tok); err != nil {
		t.Errorf("HS256 validate: %v", err)
	}
}

func TestNewValidator_OnlyJWKSURL_RS256(t *testing.T) {
	priv, pub := rsaKeyPair(t)
	kid := "k1"
	srv := httptest.NewServer(jwksServer(pub, kid))
	defer srv.Close()

	t.Setenv("GATEWAY_JWT_SECRET", "")
	t.Setenv("GATEWAY_JWT_JWKS_URL", srv.URL)

	v, err := auth.NewValidator(noopLog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok := makeRS256Token(t, priv, kid, "u", futureExp(), nil)
	if _, err := v.Validate(tok); err != nil {
		t.Errorf("RS256 validate: %v", err)
	}
}

func TestNewValidator_BothSet_JWKSPrecedence(t *testing.T) {
	priv, pub := rsaKeyPair(t)
	kid := "k1"
	srv := httptest.NewServer(jwksServer(pub, kid))
	defer srv.Close()

	t.Setenv("GATEWAY_JWT_SECRET", "ignored-secret")
	t.Setenv("GATEWAY_JWT_JWKS_URL", srv.URL)

	v, err := auth.NewValidator(noopLog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JWKS validator — RS256 token must succeed.
	tok := makeRS256Token(t, priv, kid, "u", futureExp(), nil)
	if _, err := v.Validate(tok); err != nil {
		t.Errorf("RS256 validate with JWKS precedence: %v", err)
	}
	// HS256 token must fail (wrong algorithm for JWKS validator).
	hs256tok := makeHS256Token(t, "ignored-secret", "u", futureExp(), nil)
	if _, err := v.Validate(hs256tok); err == nil {
		t.Error("expected HS256 token to fail against JWKS validator")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func futureExp() time.Time { return time.Now().Add(time.Hour) }
func pastExp() time.Time   { return time.Now().Add(-time.Hour) }

func makeHS256Token(t *testing.T, secret, sub string, exp time.Time, roles []string) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": sub, "exp": exp.Unix()}
	if len(roles) > 0 {
		claims["roles"] = roles
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign hs256: %v", err)
	}
	return signed
}

func makeRS256Token(t *testing.T, key *rsa.PrivateKey, kid, sub string, exp time.Time, roles []string) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": sub, "exp": exp.Unix()}
	if len(roles) > 0 {
		claims["roles"] = roles
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign rs256: %v", err)
	}
	return signed
}

func rsaKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen rsa key: %v", err)
	}
	return priv, &priv.PublicKey
}

// jwksServer returns an http.Handler that serves a JWKS with one RSA key.
func jwksServer(pub *rsa.PublicKey, kid string) http.Handler {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	body := fmt.Sprintf(
		`{"keys":[{"kty":"RSA","use":"sig","alg":"RS256","kid":%q,"n":%q,"e":%q}]}`,
		kid, n, e,
	)
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func noopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rr.Code != want {
		t.Errorf("HTTP status: got %d, want %d", rr.Code, want)
	}
}

func assertVendorErrorCT(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	ct := rr.Header().Get("Content-Type")
	if ct != "application/vnd.yaagents.error+json" {
		t.Errorf("Content-Type: got %q, want application/vnd.yaagents.error+json", ct)
	}
}

func decodeErrorBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return m
}
