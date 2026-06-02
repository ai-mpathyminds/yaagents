// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWrite_HeadersAndStatus verifies Write() sets all three mandatory
// response attributes (status, Content-Type, X-YAAgents-Profile) for
// every one of the 10 factory methods.
func TestWrite_HeadersAndStatus(t *testing.T) {
	ar := AgenticResponse{}

	cases := []struct {
		name       string
		resp       AgenticWritable
		wantStatus int
		wantCT     string
	}{
		{
			name:       "Accepted",
			resp:       ar.Accepted(testCtx, "op-write-01"),
			wantStatus: http.StatusAccepted,
			wantCT:     "application/vnd.yaagents.operation+json",
		},
		{
			name:       "Done",
			resp:       ar.Done(testCtx, map[string]int{"count": 3}),
			wantStatus: http.StatusOK,
			wantCT:     "application/json",
		},
		{
			name:       "Created",
			resp:       ar.Created(testCtx, map[string]string{"id": "res-1"}),
			wantStatus: http.StatusCreated,
			wantCT:     "application/json",
		},
		{
			name:       "Failed",
			resp:       ar.Failed(testCtx, "internal error"),
			wantStatus: http.StatusInternalServerError,
			wantCT:     "application/vnd.yaagents.error+json",
		},
		{
			name: "ClarificationRequired",
			resp: ar.ClarificationRequired(testCtx, []RequiredInput{
				{Name: "goal", Location: "body", Type: "string", Required: true, Question: "Goal?"},
			}),
			wantStatus: http.StatusBadRequest,
			wantCT:     "application/vnd.yaagents.clarification+json",
		},
		{
			name: "ValidationFailed",
			resp: ar.ValidationFailed(testCtx, []ValidationError{
				{Field: "amount", Message: "must be positive"},
			}),
			wantStatus: http.StatusUnprocessableEntity,
			wantCT:     "application/vnd.yaagents.validation-error+json",
		},
		{
			name:       "ApprovalRequired",
			resp:       ar.ApprovalRequired(testCtx, []string{"boss@co"}, "large budget"),
			wantStatus: http.StatusPreconditionFailed,
			wantCT:     "application/vnd.yaagents.approval-required+json",
		},
		{
			name:       "Forbidden",
			resp:       ar.Forbidden(testCtx, "not allowed"),
			wantStatus: http.StatusForbidden,
			wantCT:     "application/vnd.yaagents.error+json",
		},
		{
			name:       "Conflict",
			resp:       ar.Conflict(testCtx, "duplicate key"),
			wantStatus: http.StatusConflict,
			wantCT:     "application/vnd.yaagents.conflict+json",
		},
		{
			name:       "FailedDependency",
			resp:       ar.FailedDependency(testCtx, "payment-svc", "unavailable"),
			wantStatus: http.StatusFailedDependency,
			wantCT:     "application/vnd.yaagents.error+json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if err := Write(rec, tc.resp); err != nil {
				t.Fatalf("Write() error: %v", err)
			}

			if got := rec.Code; got != tc.wantStatus {
				t.Errorf("status = %d, want %d", got, tc.wantStatus)
			}
			if got := rec.Header().Get("Content-Type"); got != tc.wantCT {
				t.Errorf("Content-Type = %q, want %q", got, tc.wantCT)
			}
			if got := rec.Header().Get("X-YAAgents-Profile"); got != ProfileVersion {
				t.Errorf("X-YAAgents-Profile = %q, want %q", got, ProfileVersion)
			}
			if rec.Body.Len() == 0 {
				t.Error("response body is empty")
			}
		})
	}
}

// TestWrite_ProfileHeader_AllTen explicitly asserts X-YAAgents-Profile: v0.3
// for each of the 10 factory methods (belt-and-suspenders per acceptance criteria).
func TestWrite_ProfileHeader_AllTen(t *testing.T) {
	ar := AgenticResponse{}
	resps := []AgenticWritable{
		ar.Accepted(testCtx, "op-ph"),
		ar.Done(testCtx, struct{}{}),
		ar.Created(testCtx, struct{}{}),
		ar.Failed(testCtx, "err"),
		ar.ClarificationRequired(testCtx, []RequiredInput{
			{Name: "x", Location: "body", Type: "string", Required: true, Question: "q?"},
		}),
		ar.ValidationFailed(testCtx, []ValidationError{{Field: "f", Message: "m"}}),
		ar.ApprovalRequired(testCtx, nil, "reason"),
		ar.Forbidden(testCtx, "no"),
		ar.Conflict(testCtx, "dup"),
		ar.FailedDependency(testCtx, "dep", "down"),
	}
	if len(resps) != 10 {
		t.Fatalf("expected 10 responses, got %d", len(resps))
	}
	for i, resp := range resps {
		rec := httptest.NewRecorder()
		if err := Write(rec, resp); err != nil {
			t.Fatalf("resp[%d] Write() error: %v", i, err)
		}
		if got := rec.Header().Get("X-YAAgents-Profile"); got != "v0.3" {
			t.Errorf("resp[%d] X-YAAgents-Profile = %q, want v0.3", i, got)
		}
	}
}

// TestWrite_BodyContents verifies Write actually flushes body bytes to the recorder.
func TestWrite_BodyContents(t *testing.T) {
	ar := AgenticResponse{}
	rec := httptest.NewRecorder()
	if err := Write(rec, ar.Done(testCtx, map[string]string{"hello": "world"})); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("body is empty")
	}
	if !strings.Contains(body, "hello") {
		t.Errorf("body %q does not contain 'hello'", body)
	}
}
