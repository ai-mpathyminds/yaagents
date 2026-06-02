// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package echo is the labstack/echo/v4 adapter for the YAAgents Go SDK.
// It re-exports the core sdkgo API and adds URLParam backed by echo's
// path-parameter extraction via a request-context bridge middleware.
//
// Register InjectContext() on the router before registering handlers:
//
//	import (
//	    "github.com/labstack/echo/v4"
//	    echoadapter "github.com/ai-mpathyminds/yaagents-sdk-go/adapters/echo"
//	)
//
//	e := echo.New()
//	e.Use(echoadapter.InjectContext())
//	e.POST("/campaigns/:id/optimizations", func(c echo.Context) error {
//	    ctx := echoadapter.FromRequest(c.Request())
//	    id  := echoadapter.URLParam(c.Request(), "id")
//	    ar  := echoadapter.AgenticResponse{}
//	    return echoadapter.Write(c.Response(), ar.Created(ctx, payload))
//	})
package echo

import (
	"context"
	"net/http"

	goecho "github.com/labstack/echo/v4"

	sdkgo "github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

// echoCtxKey is the unexported context key for echo.Context storage.
type echoCtxKey struct{}

// InjectContext returns an echo middleware that stores the echo.Context inside
// the request context. This is required for URLParam to retrieve path
// parameters from a plain *http.Request.
func InjectContext() goecho.MiddlewareFunc {
	return func(next goecho.HandlerFunc) goecho.HandlerFunc {
		return func(c goecho.Context) error {
			c.SetRequest(c.Request().WithContext(
				context.WithValue(c.Request().Context(), echoCtxKey{}, c),
			))
			return next(c)
		}
	}
}

// URLParam extracts a URL path parameter by name from an echo-routed request.
// Requires InjectContext() middleware to be registered on the router.
// Returns an empty string when the parameter is absent or middleware is inactive.
func URLParam(r *http.Request, name string) string {
	c, ok := r.Context().Value(echoCtxKey{}).(goecho.Context)
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
// Pass c.Response() (*echo.Response) as w — it satisfies http.ResponseWriter.
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
