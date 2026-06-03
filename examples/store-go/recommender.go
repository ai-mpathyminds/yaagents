// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package main

import "fmt"

// mockRecommend returns same-category products as recommendations.
//
// Replace this function with a real LLM call to build a genuine
// AI-powered recommender. See examples/store/skills/ for starter prompts
// (they apply equally to the Python and Go implementations).
func mockRecommend(
	products []Product,
	seed *Product,
	customer *Customer,
	limit int,
	excludePurchased bool,
) ([]Product, string) {
	if limit <= 0 {
		limit = 3
	}

	// Build purchased set for filtering.
	purchased := map[string]bool{}
	if customer != nil && excludePurchased {
		for _, pid := range customer.Purchased {
			purchased[pid] = true
		}
	}

	// Collect same-category candidates.
	var candidates []Product
	for _, p := range products {
		if p.Category == seed.Category && p.ID != seed.ID && !purchased[p.ID] {
			candidates = append(candidates, p)
		}
	}

	// Truncate to limit.
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	reason := fmt.Sprintf("Same category as '%s' (%s)", seed.Name, seed.Category)
	if customer != nil {
		reason += fmt.Sprintf("; personalised for customer %s (purchased items excluded)", customer.ID)
	}
	return candidates, reason
}
