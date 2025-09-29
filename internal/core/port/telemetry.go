package port

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryProbe defines the interface for domain telemetry
// This allows the domain to emit telemetry events without knowing the implementation
type Telemetry interface {
	// Tracing - Span creation
	StartRepositorySpan(ctx context.Context, operation string, entity string, attrs []attribute.KeyValue) (context.Context, trace.Span)
	StartServiceSpan(ctx context.Context, service string, operation string, userID int, attrs []attribute.KeyValue) (context.Context, trace.Span)
	StartHTTPSpan(ctx context.Context, method string, path string, attrs []attribute.KeyValue) (context.Context, trace.Span)

	// Repository operations
	RecordRepositoryOperation(ctx context.Context, operation string, entity string, duration time.Duration, err error)
	RecordRepositoryQuery(ctx context.Context, operation string, entity string, query string, args []interface{})

	// Service operations
	RecordServiceOperation(ctx context.Context, service string, operation string, userID int, duration time.Duration, err error)

	// Business events
	RecordBusinessEvent(ctx context.Context, event string, entity string, entityID string, userID int, metadata map[string]interface{})

	// HTTP operations
	RecordHTTPOperation(ctx context.Context, method string, path string, statusCode int, duration time.Duration)

	// Errors
	RecordError(ctx context.Context, operation string, err error, metadata map[string]interface{})
}
