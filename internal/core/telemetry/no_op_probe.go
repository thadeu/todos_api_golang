package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"todoapp/internal/core/port"
)

// NoOpProbe is a probe that does nothing - useful for testing or when telemetry is disabled
type NoOpProbe struct{}

func NewNoOpProbe() port.Telemetry {
	return &NoOpProbe{}
}

// Tracing methods - return no-op span
func (p *NoOpProbe) StartRepositorySpan(ctx context.Context, operation string, entity string, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx) // Return existing span (no-op)
}

func (p *NoOpProbe) StartServiceSpan(ctx context.Context, service string, operation string, userID int, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx) // Return existing span (no-op)
}

func (p *NoOpProbe) StartHTTPSpan(ctx context.Context, method string, path string, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx) // Return existing span (no-op)
}

func (p *NoOpProbe) RecordRepositoryOperation(ctx context.Context, operation string, entity string, duration time.Duration, err error) {
	// No operation
}

func (p *NoOpProbe) RecordRepositoryQuery(ctx context.Context, operation string, entity string, query string, args []interface{}) {
	// No operation
}

func (p *NoOpProbe) RecordServiceOperation(ctx context.Context, service string, operation string, userID int, duration time.Duration, err error) {
	// No operation
}

func (p *NoOpProbe) RecordBusinessEvent(ctx context.Context, event string, entity string, entityID string, userID int, metadata map[string]interface{}) {
	// No operation
}

func (p *NoOpProbe) RecordHTTPOperation(ctx context.Context, method string, path string, statusCode int, duration time.Duration) {
	// No operation
}

func (p *NoOpProbe) RecordError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	// No operation
}

// TelemetryOperation is a helper to measure operation duration
type TelemetryOperation struct {
	probe     port.Telemetry
	ctx       context.Context
	startTime time.Time
	operation string
	entity    string
}

// StartOperation begins measuring an operation
func StartOperation(probe port.Telemetry, ctx context.Context, operation, entity string) *TelemetryOperation {
	return &TelemetryOperation{
		probe:     probe,
		ctx:       ctx,
		startTime: time.Now(),
		operation: operation,
		entity:    entity,
	}
}

// End marks the operation as completed
func (op *TelemetryOperation) End(err error) {
	if op.probe != nil {
		duration := time.Since(op.startTime)
		op.probe.RecordRepositoryOperation(op.ctx, op.operation, op.entity, duration, err)
	}
}
