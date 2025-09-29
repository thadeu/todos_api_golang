package telemetry

import (
	"context"
	"time"

	"todoapp/internal/core/port"
)

// NoOpProbe implements Telemetry with no operations - useful for testing or when telemetry is disabled
type NoOpProbe struct{}

func NewNoOpProbe() port.Telemetry {
	return &NoOpProbe{}
}

// NoOpSpan implements the Span interface with no operations
type NoOpSpan struct{}

func (s *NoOpSpan) End()                                       {}
func (s *NoOpSpan) SetAttributes(attrs map[string]interface{}) {}
func (s *NoOpSpan) SetStatus(code string, message string)      {}
func (s *NoOpSpan) RecordError(err error)                      {}

// Tracing methods - return no-op span
func (p *NoOpProbe) StartRepositorySpan(ctx context.Context, operation string, entity string, attrs map[string]interface{}) (context.Context, port.Span) {
	return ctx, &NoOpSpan{}
}

func (p *NoOpProbe) StartServiceSpan(ctx context.Context, service string, operation string, userID int, attrs map[string]interface{}) (context.Context, port.Span) {
	return ctx, &NoOpSpan{}
}

func (p *NoOpProbe) StartHTTPSpan(ctx context.Context, method string, path string, attrs map[string]interface{}) (context.Context, port.Span) {
	return ctx, &NoOpSpan{}
}

// Repository operations
func (p *NoOpProbe) RecordRepositoryOperation(ctx context.Context, operation string, entity string, duration time.Duration, err error) {
	// No operation
}

func (p *NoOpProbe) RecordRepositoryQuery(ctx context.Context, operation string, entity string, query string, args []interface{}) {
	// No operation
}

// Service operations
func (p *NoOpProbe) RecordServiceOperation(ctx context.Context, service string, operation string, userID int, duration time.Duration, err error) {
	// No operation
}

// Business events
func (p *NoOpProbe) RecordBusinessEvent(ctx context.Context, event string, entity string, entityID string, userID int, metadata map[string]interface{}) {
	// No operation
}

// HTTP operations
func (p *NoOpProbe) RecordHTTPOperation(ctx context.Context, method string, path string, statusCode int, duration time.Duration) {
	// No operation
}

// Errors
func (p *NoOpProbe) RecordError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	// No operation
}
