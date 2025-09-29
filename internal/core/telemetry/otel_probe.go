package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"todos/internal/core/port"
)

// OTELProbe implements Telemetry using OpenTelemetry
type OTELProbe struct {
	logger *slog.Logger
}

func NewOTELProbe(logger *slog.Logger) port.Telemetry {
	return &OTELProbe{
		logger: logger,
	}
}

// OTelSpan wraps OpenTelemetry span to implement our generic Span interface
type OTelSpan struct {
	span trace.Span
}

func (s *OTelSpan) End() {
	s.span.End()
}

func (s *OTelSpan) SetAttributes(attrs map[string]interface{}) {
	var otelAttrs []attribute.KeyValue
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			otelAttrs = append(otelAttrs, attribute.String(key, v))
		case int:
			otelAttrs = append(otelAttrs, attribute.Int(key, v))
		case int64:
			otelAttrs = append(otelAttrs, attribute.Int64(key, v))
		case float64:
			otelAttrs = append(otelAttrs, attribute.Float64(key, v))
		case bool:
			otelAttrs = append(otelAttrs, attribute.Bool(key, v))
		default:
			otelAttrs = append(otelAttrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
	s.span.SetAttributes(otelAttrs...)
}

func (s *OTelSpan) SetStatus(code string, message string) {
	var statusCode codes.Code
	switch code {
	case "ok":
		statusCode = codes.Ok
	case "error":
		statusCode = codes.Error
	default:
		statusCode = codes.Unset
	}
	s.span.SetStatus(statusCode, message)
}

func (s *OTelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

// Tracing methods
func (p *OTELProbe) StartRepositorySpan(ctx context.Context, operation string, entity string, attrs map[string]interface{}) (context.Context, port.Span) {
	spanName := fmt.Sprintf("repository.%s.%s", entity, operation)

	// Add standard repository attributes
	standardAttrs := []attribute.KeyValue{
		attribute.String("repository.entity", entity),
		attribute.String("repository.operation", operation),
		attribute.String("component", "repository"),
	}

	// Convert map attrs to OpenTelemetry attributes
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			standardAttrs = append(standardAttrs, attribute.String(key, v))
		case int:
			standardAttrs = append(standardAttrs, attribute.Int(key, v))
		case int64:
			standardAttrs = append(standardAttrs, attribute.Int64(key, v))
		case float64:
			standardAttrs = append(standardAttrs, attribute.Float64(key, v))
		case bool:
			standardAttrs = append(standardAttrs, attribute.Bool(key, v))
		default:
			standardAttrs = append(standardAttrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	ctx, span := otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(standardAttrs...))
	return ctx, &OTelSpan{span: span}
}

func (p *OTELProbe) StartServiceSpan(ctx context.Context, service string, operation string, userID int, attrs map[string]interface{}) (context.Context, port.Span) {
	spanName := fmt.Sprintf("service.%s.%s", service, operation)

	// Add standard service attributes
	standardAttrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("service.operation", operation),
		attribute.Int("user.id", userID),
		attribute.String("component", "service"),
	}

	// Convert map attrs to OpenTelemetry attributes
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			standardAttrs = append(standardAttrs, attribute.String(key, v))
		case int:
			standardAttrs = append(standardAttrs, attribute.Int(key, v))
		case int64:
			standardAttrs = append(standardAttrs, attribute.Int64(key, v))
		case float64:
			standardAttrs = append(standardAttrs, attribute.Float64(key, v))
		case bool:
			standardAttrs = append(standardAttrs, attribute.Bool(key, v))
		default:
			standardAttrs = append(standardAttrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	ctx, span := otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(standardAttrs...))
	return ctx, &OTelSpan{span: span}
}

func (p *OTELProbe) StartHTTPSpan(ctx context.Context, method string, path string, attrs map[string]interface{}) (context.Context, port.Span) {
	spanName := fmt.Sprintf("http.%s", path)

	// Add standard HTTP attributes
	standardAttrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("component", "http"),
	}

	// Convert map attrs to OpenTelemetry attributes
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			standardAttrs = append(standardAttrs, attribute.String(key, v))
		case int:
			standardAttrs = append(standardAttrs, attribute.Int(key, v))
		case int64:
			standardAttrs = append(standardAttrs, attribute.Int64(key, v))
		case float64:
			standardAttrs = append(standardAttrs, attribute.Float64(key, v))
		case bool:
			standardAttrs = append(standardAttrs, attribute.Bool(key, v))
		default:
			standardAttrs = append(standardAttrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	ctx, span := otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(standardAttrs...))
	return ctx, &OTelSpan{span: span}
}

// Repository operations
func (p *OTELProbe) RecordRepositoryOperation(ctx context.Context, operation string, entity string, duration time.Duration, err error) {
	// Get current span from context (created by StartRepositorySpan)
	span := trace.SpanFromContext(ctx)

	// Add operation details to existing span
	span.SetAttributes(
		attribute.String("operation", operation),
		attribute.String("entity", entity),
		attribute.Int64("duration_ns", duration.Nanoseconds()),
		attribute.Bool("has_error", err != nil),
	)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "Repository operation failed",
			"operation", operation,
			"entity", entity,
			"duration_ns", duration.Nanoseconds(),
			"error", err)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (p *OTELProbe) RecordRepositoryQuery(ctx context.Context, operation string, entity string, query string, args []interface{}) {
	// Log the query for debugging (but not the full args for security)
	safeArgs := make([]string, len(args))
	for i := range args {
		safeArgs[i] = fmt.Sprintf("%T", args[i])
	}

	p.logger.DebugContext(ctx, "Executing repository query",
		"operation", operation,
		"entity", entity,
		"query", query,
		"args_types", safeArgs)
}

// Service operations
func (p *OTELProbe) RecordServiceOperation(ctx context.Context, service string, operation string, userID int, duration time.Duration, err error) {
	// Get current span from context (created by StartServiceSpan)
	span := trace.SpanFromContext(ctx)

	// Add operation details to existing span
	span.SetAttributes(
		attribute.String("service", service),
		attribute.String("operation", operation),
		attribute.Int("user_id", userID),
		attribute.Int64("duration_ns", duration.Nanoseconds()),
		attribute.Bool("has_error", err != nil),
	)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		p.logger.ErrorContext(ctx, "Service operation failed",
			"service", service,
			"operation", operation,
			"user_id", userID,
			"duration_ns", duration.Nanoseconds(),
			"error", err)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// Business events
func (p *OTELProbe) RecordBusinessEvent(ctx context.Context, event string, entity string, entityID string, userID int, metadata map[string]interface{}) {
	// Create a span for the business event
	ctx, span := p.StartRepositorySpan(ctx, fmt.Sprintf("event.%s", event), entity, map[string]interface{}{
		"event":     event,
		"entity":    entity,
		"entity_id": entityID,
		"user_id":   userID,
	})

	// Add metadata as span attributes
	for key, value := range metadata {
		switch v := value.(type) {
		case string:
			span.SetAttributes(map[string]interface{}{key: v})
		case int:
			span.SetAttributes(map[string]interface{}{key: v})
		case int64:
			span.SetAttributes(map[string]interface{}{key: v})
		case float64:
			span.SetAttributes(map[string]interface{}{key: v})
		case bool:
			span.SetAttributes(map[string]interface{}{key: v})
		default:
			span.SetAttributes(map[string]interface{}{key: fmt.Sprintf("%v", v)})
		}
	}

	span.End()

	p.logger.InfoContext(ctx, "Business event recorded",
		"event", event,
		"entity", entity,
		"entity_id", entityID,
		"user_id", userID,
		"metadata", metadata)
}

// HTTP operations
func (p *OTELProbe) RecordHTTPOperation(ctx context.Context, method string, path string, statusCode int, duration time.Duration) {
	// Get current span from context (created by StartHTTPSpan)
	span := trace.SpanFromContext(ctx)

	// Add HTTP operation details to existing span
	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.duration_ns", duration.Nanoseconds()),
	)

	// Set span status based on HTTP status
	if statusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	p.logger.InfoContext(ctx, "HTTP operation completed",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ns", duration.Nanoseconds())
}

// Errors
func (p *OTELProbe) RecordError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	p.logger.ErrorContext(ctx, "Operation error recorded",
		"operation", operation,
		"error", err,
		"metadata", metadata)
}
