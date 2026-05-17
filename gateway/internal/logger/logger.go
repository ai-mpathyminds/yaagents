// Package logger provides a structured JSON logger for the yaagents gateway.
//
// Per-request fields (request_id, correlation_id) are added by callers via
// [slog.Logger.With]:
//
//	reqLog := log.With(
//	    slog.String("request_id", reqID),
//	    slog.String("correlation_id", corrID),
//	)
package logger

import (
	"log/slog"
	"os"
)

// New returns a *slog.Logger that emits structured JSON to stdout.
// The log level is INFO; callers inject per-request context via [slog.Logger.With].
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
