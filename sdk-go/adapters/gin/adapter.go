// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package gin is the gin-gonic/gin adapter for the YAAgents Go SDK.
// It re-exports the core sdkgo API and adds URLParam backed by gin's
// path-parameter extraction via a request-context bridge middleware.
//
// Register InjectContext() on the router before registering handlers:
//
//	import (
//	    "github.com/gin-gonic/gin"
//	    ginadapter "github.com/ai-mpathyminds/yaagents-sdk-go/adapters/gin"
//	)
//
//	r := gin.New()
//	r.Use(ginadapter.InjectContext())
//	r.POST("/campaigns/:id/optimizations", func(c *gin.Context) {
//	    ctx := ginadapter.FromRequest(c.Request)
//	    id  := ginadapter.URLParam(c.Request, "id")
//	    ar  := ginadapter.AgenticResponse{}
//	    ginadapter.Write(c.Writer, ar.Created(ctx, payload))
//	})
package gin

import (
	"context"
	"net/http"

	gogin "github.com/gin-gonic/gin"

	sdkgo "github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

// ginCtxKey is the unexported context key for *gin.Context storage.
type ginCtxKey struct{}

// InjectContext returns a gin middleware that stores the *gin.Context inside
// the request context. This is required for URLParam to retrieve path
// parameters from a plain *http.Request.
func InjectContext() gogin.HandlerFunc {
	return func(c *gogin.Context) {
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), ginCtxKey{}, c),
		)
		c.Next()
	}
}

// URLParam extracts a URL path parameter by name from a gin-routed request.
// Requires InjectContext() middleware to be registered on the router.
// Returns an empty string when the parameter is absent or middleware is inactive.
func URLParam(r *http.Request, name string) string {
	c, ok := r.Context().Value(ginCtxKey{}).(*gogin.Context)
	if !ok {
		return ""
	}
	return c.Param(name)
}

// FromRequest extracts the AgenticContext from gateway-injected request headers.
// Delegates to sdkgo.FromRequest.
func FromRequest(r *http.Request) sdkgo.AgenticContext { return sdkgo.FromRequest(r) }

// Write serializes resp to w, setting HTTP status, Content-Type, and
// X-YAAgents-Profile: v0.3. Delegates to sdkgo.Write.
// Pass c.Writer (gin.ResponseWriter) as w — it satisfies http.ResponseWriter.
func Write(w http.ResponseWriter, resp sdkgo.AgenticWritable) error {
	return sdkgo.Write(w, resp)
}

// Type aliases — importing this package gives access to the full SDK surface.
type (
	AgenticContext  = sdkgo.AgenticContext
	AgenticResponse = sdkgo.AgenticResponse
	AgenticWritable = sdkgo.AgenticWritable
	RequiredInput   = sdkgo.RequiredInput
	ValidationError = sdkgo.ValidationError
)
