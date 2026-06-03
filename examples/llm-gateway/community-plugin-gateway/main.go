// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Command community-plugin-gateway verifies that an operator-authored community
// plugin can be compiled into a custom yaagents gateway binary (PRD §6.6).
//
// This binary is intentionally minimal: it imports the community-example plugin
// as a side-effect (triggering its init() registration), then prints all
// registered plugin names and exits 0.
//
// In a production scenario an operator would instead copy the full gateway
// main.go here, add the blank import for their plugin, and rebuild — the
// gateway loader picks it up automatically via plugin.Registered().
//
// Verification in test-e2e.sh:
//
//	cd examples/llm-gateway/community-plugin-gateway
//	go build -o community-gw-verify .
//	./community-gw-verify
//	# → output contains "plugin registered: community-example"
package main

import (
	"fmt"

	"github.com/ai-mpathyminds/yaagents-gateway/plugin"

	// Side-effect import: registers the community-example plugin into the global
	// registry (plugin.Register is called from communityplugin's init function).
	_ "github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin"
)

func main() {
	for _, p := range plugin.Registered() {
		fmt.Printf("plugin registered: %s\n", p.Name())
	}
}
