package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type TraceHandler struct {
	slog.Handler
}

// Handle injects OpenTelemetry traceId and spanId into the log record
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		r.AddAttrs(slog.String("traceId", spanContext.TraceID().String()))
	}
	if spanContext.HasSpanID() {
		r.AddAttrs(slog.String("spanId", spanContext.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

// NewLogger creates a new JSON logger with Trace Correlation enabled
func NewLogger() *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	return slog.New(&TraceHandler{Handler: baseHandler})
}
