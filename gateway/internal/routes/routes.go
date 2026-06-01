// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package routes parses and validates the gateway route configuration (routes.yaml).
//
// Schema (PRD §5.4):
//
//	routes:
//	  - id: <string, required, unique>
//	    method: <HTTP verb, required>
//	    path: <string, required; {param} placeholders allowed>
//	    target: <URL, required>
//	    roles: [<string>, ...]   # optional; empty = open to all authenticated callers
//	    tenantRequired: <bool>   # optional, default false
//	    audit: <bool>            # optional, default false
package routes

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// validMethods is the set of allowed HTTP methods (upper-case canonical form).
var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

// placeholderRe matches {name} tokens inside a route path.
var placeholderRe = regexp.MustCompile(`\{([^{}]*)\}`)

// Route is a single validated gateway route parsed from routes.yaml.
type Route struct {
	ID             string   `yaml:"id"`
	Method         string   `yaml:"method"`
	Path           string   `yaml:"path"`
	Target         string   `yaml:"target"`
	Roles          []string `yaml:"roles"`
	TenantRequired bool     `yaml:"tenantRequired"`
	Audit          bool     `yaml:"audit"`

	// Mode specifies the proxy mode for this route.
	// "" (empty, default) = standard httputil.ReverseProxy (GW-4 path).
	// "sse" = SSE-aware pipe-and-flush proxy (LLM-1; activates internal/llm).
	Mode string `yaml:"mode"`

	// ExecutionTimeoutSeconds caps the upstream call duration (LLM-3).
	// 0 (default) = no timeout.
	// SSE routes: deadline = now + ExecutionTimeoutSeconds + 30 s (PRD §7.1 SSEReadTimeout).
	// Non-SSE routes: deadline = now + ExecutionTimeoutSeconds.
	// Exceeded → 500 application/vnd.yaagents.error+json code: "EXECUTION_TIMEOUT".
	ExecutionTimeoutSeconds int `yaml:"executionTimeoutSeconds"`

	// Plugins holds per-route plugin overrides (PRD §5.4.2). A plugin listed
	// here may set enabled: false to bypass it for this route only. Disabling
	// token-validator per-route is a fatal boot error (ADR PI2-yaa-0001 §5).
	// Example YAML:
	//   plugins:
	//     tenant-injector:
	//       enabled: false
	Plugins map[string]map[string]any `yaml:"plugins"`

	// PathParams holds the ordered placeholder names extracted from Path
	// (e.g. ["campaignId"] for /campaigns/{campaignId}/optimizations).
	// Populated by Load; not present in the YAML source.
	PathParams []string `yaml:"-"`
}

// routeFile is the top-level envelope of routes.yaml.
type routeFile struct {
	Routes []Route `yaml:"routes"`
}

// Load reads, parses and validates the route configuration at path.
// It returns a non-nil error — and exits with a descriptive message — when:
//   - the file cannot be read
//   - the YAML is malformed
//   - any route fails schema validation
//
// Callers should treat a non-nil error as a fatal boot failure (fail-fast).
func Load(path string) ([]Route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading routes file %q: %w", path, err)
	}
	return parse(data)
}

// parse decodes YAML bytes and delegates to validate.
func parse(data []byte) ([]Route, error) {
	var rf routeFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parsing routes YAML: %w", err)
	}
	if err := validate(rf.Routes); err != nil {
		return nil, err
	}
	return rf.Routes, nil
}

// validate checks every route for required fields, valid values, and uniqueness.
// It also populates Route.PathParams in-place.
func validate(routes []Route) error {
	seen := make(map[string]bool, len(routes))
	var errs []string

	for i := range routes {
		r := &routes[i]
		label := fmt.Sprintf("route[%d]", i)

		// id — required, unique.
		switch {
		case r.ID == "":
			errs = append(errs, label+": id is required")
		case seen[r.ID]:
			errs = append(errs, fmt.Sprintf("route id %q is duplicated", r.ID))
		default:
			seen[r.ID] = true
			label = fmt.Sprintf("route %q", r.ID)
		}

		// method — required, valid HTTP verb.
		if r.Method == "" {
			errs = append(errs, label+": method is required")
		} else if !validMethods[strings.ToUpper(r.Method)] {
			errs = append(errs, fmt.Sprintf("%s: method %q is not a valid HTTP method", label, r.Method))
		} else {
			r.Method = strings.ToUpper(r.Method)
		}

		// path — required; extract {param} placeholders.
		if r.Path == "" {
			errs = append(errs, label+": path is required")
		} else {
			params, perr := parsePlaceholders(r.Path)
			if perr != nil {
				errs = append(errs, fmt.Sprintf("%s: path: %s", label, perr.Error()))
			} else {
				r.PathParams = params
			}
		}

		// target — required, must be a parseable absolute URL.
		if r.Target == "" {
			errs = append(errs, label+": target is required")
		} else if _, uerr := url.ParseRequestURI(r.Target); uerr != nil {
			errs = append(errs, fmt.Sprintf("%s: target %q is not a valid URL: %s", label, r.Target, uerr.Error()))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid route configuration:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// parsePlaceholders extracts {name} tokens from a route path.
// Returns an error for empty placeholders ({}) or duplicate names.
func parsePlaceholders(path string) ([]string, error) {
	matches := placeholderRe.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))
	seen := make(map[string]bool, len(matches))

	for _, m := range matches {
		name := m[1]
		if name == "" {
			return nil, errors.New("empty placeholder {} not allowed")
		}
		if seen[name] {
			return nil, fmt.Errorf("placeholder {%s} appears more than once", name)
		}
		seen[name] = true
		names = append(names, name)
	}
	return names, nil
}
