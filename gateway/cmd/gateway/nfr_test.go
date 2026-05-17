// Package main — NFR-GW-1 static checks for secret hygiene.
//
// These tests run as part of the normal `go test ./...` suite.  They enforce
// the properties that trivy config scan and the REL-6 CI grep gate verify at
// publish time, so regressions are caught locally on every `go test` run:
//
//  1. docker/gateway/Dockerfile contains no ENV instruction that sets a secret.
//  2. docker/gateway/Dockerfile EXPOSEs the canonical yaagents-gateway port (8120).
//  3. No .env file exists anywhere under the gateway module root.
//
// Path convention: tests in cmd/gateway/ resolve paths relative to the package
// directory, which Go guarantees to be yaagents/gateway/cmd/gateway/ at test time.
// Three "../" navigate to the yaagents/ repo root.
package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repoRoot returns the yaagents/ repo root, three directories above the
// cmd/gateway/ package (yaagents/gateway/cmd/gateway/ → yaagents/).
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	// wd = .../yaagents/gateway/cmd/gateway
	return filepath.Join(wd, "..", "..", "..")
}

// gatewayDockerfile returns the absolute path to docker/gateway/Dockerfile.
func gatewayDockerfile(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "docker", "gateway", "Dockerfile")
}

// secretENVRe matches any Dockerfile ENV instruction whose value contains a
// credential placeholder.  Pattern covers:
//
//	ENV GATEWAY_JWT_SECRET=<value>
//	ENV SOME_TOKEN=abc
//	ENV DB_PASSWORD=hunter2
//
// A bare comment line beginning with "#" is not an ENV instruction and is
// explicitly excluded by the regexp anchoring on "ENV".
var secretENVRe = regexp.MustCompile(`(?i)^\s*ENV\s+\S*(SECRET|TOKEN|PASSWORD|API_KEY)\S*\s*=`)

// TestDockerfileNoSecretEnv verifies that docker/gateway/Dockerfile contains
// no ENV instruction whose name includes SECRET, TOKEN, PASSWORD, or API_KEY.
// This is the local counterpart of the REL-6 CI grep gate (NFR-GW-1 AC).
func TestDockerfileNoSecretEnv(t *testing.T) {
	path := gatewayDockerfile(t)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot open Dockerfile at %s: %v", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if secretENVRe.MatchString(line) {
			t.Errorf("Dockerfile line %d contains a secret ENV instruction (NFR-GW-1): %s", lineNum, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning Dockerfile: %v", err)
	}
}

// TestDockerfileExposePort verifies that the Dockerfile EXPOSEs 8120 — the
// canonical yaagents-gateway port in the portfolio port allocation table.
func TestDockerfileExposePort(t *testing.T) {
	path := gatewayDockerfile(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}

	const wantPort = "EXPOSE 8120"
	if !strings.Contains(string(data), wantPort) {
		t.Errorf("Dockerfile does not contain %q; found:\n%s", wantPort, data)
	}
}

// TestDockerfileNoJWTSecretDefault verifies that GATEWAY_JWT_SECRET does not
// appear as a left-hand side of any ENV assignment (NFR-GW-1 AC 1).
func TestDockerfileNoJWTSecretDefault(t *testing.T) {
	path := gatewayDockerfile(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}

	// A substring search is sufficient: "GATEWAY_JWT_SECRET=" would only appear
	// in an ENV instruction or in a comment.  Comments must not set the value either.
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue // skip comments
		}
		if strings.Contains(line, "GATEWAY_JWT_SECRET=") {
			t.Errorf("Dockerfile contains GATEWAY_JWT_SECRET= outside a comment: %q", line)
		}
	}
}

// TestNoEnvFileInGateway verifies that no .env file (or .env.* variant) exists
// under the gateway module directory.  Such files must never be committed
// (NFR-GW-1; .gitignore gate is belt-and-suspenders).
func TestNoEnvFileInGateway(t *testing.T) {
	// Gateway module root is two directories up from cmd/gateway/.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	gatewayRoot := filepath.Join(wd, "..", "..")

	err = filepath.WalkDir(gatewayRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		name := d.Name()
		if name == ".env" || strings.HasPrefix(name, ".env.") {
			t.Errorf("found .env file in gateway module tree (must not be committed): %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking gateway root: %v", err)
	}
}
