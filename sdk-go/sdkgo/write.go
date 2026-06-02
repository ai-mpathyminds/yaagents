// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo

import "net/http"

// Write serializes resp to w. It sets:
//   - Content-Type  → resp.ContentType()
//   - X-YAAgents-Profile → ProfileVersion ("v0.3")
//   - HTTP status   → resp.Status()
//
// Returns any serialization or write error. Callers should not call
// w.WriteHeader after Write returns.
func Write(w http.ResponseWriter, resp AgenticWritable) error {
	b, err := resp.Body()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", resp.ContentType())
	w.Header().Set("X-YAAgents-Profile", ProfileVersion)
	w.WriteHeader(resp.Status())
	_, err = w.Write(b)
	return err
}
