// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo_test

import (
	"testing"

	"github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

func TestProfileVersion(t *testing.T) {
	const want = "v0.3"
	if sdkgo.ProfileVersion != want {
		t.Errorf("ProfileVersion = %q, want %q", sdkgo.ProfileVersion, want)
	}
}
