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

	"todoapp/internal/core/port"
)

type OTELProbe struct {
	logger *slog.Logger
}

func NewOTELProbe(logger *slog.Logger) port.Telemetry {
	return &OTELProbe{
		logger: logger,
	}
}

func (p *OTELProbe) StartRepositorySpan(ctx context.Context, operation string, entity string, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("repository.%s.%s", entity, operation)

	standardAttrs := []attribute.KeyValue{
		attribute.String("repository.entity", entity),
		attribute.String("repository.operation", operation),
		attribute.String("component", "repository"),
	}

	allAttrs := append(standardAttrs, attrs...)

	return otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(allAttrs...))
}

func (p *OTELProbe) StartServiceSpan(ctx context.Context, service string, operation string, userID int, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("service.%s.%s", service, operation)

	standardAttrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("service.operation", operation),
		attribute.Int("user.id", userID),
		attribute.String("component", "service"),
	}

	allAttrs := append(standardAttrs, attrs...)

	return otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(allAttrs...))
}

func (p *OTELProbe) StartHTTPSpan(ctx context.Context, method string, path string, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("http.%s", path)

	standardAttrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("component", "http"),
	}

	allAttrs := append(standardAttrs, attrs...)

	return otel.Tracer("todoapp").Start(ctx, spanName, trace.WithAttributes(allAttrs...))
}

func (p *OTELProbe) RecordRepositoryOperation(ctx context.Context, operation string, entity string, duration time.Duration, err error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("operation", operation),
		attribute.String("entity", entity),
		attribute.Int64("duration_ns", duration.Nanoseconds()),
		attribute.Bool("has_error", err != nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	safeArgs := make([]string, len(args))

	for i := range args {
		safeArgs[i] = fmt.Sprintf("%T", args[i])
	}

	p.logger.DebugContext(ctx, "Repository query executed",
		"operation", operation,
		"entity", entity,
		"query", query,
		"args_types", safeArgs)
}

func (p *OTELProbe) RecordServiceOperation(ctx context.Context, service string, operation string, userID int, duration time.Duration, err error) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("service", service),
		attribute.String("operation", operation),
		attribute.Int("user_id", userID),
		attribute.Int64("duration_ns", duration.Nanoseconds()),
		attribute.Bool("has_error", err != nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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

func (p *OTELProbe) RecordBusinessEvent(ctx context.Context, event string, entity string, entityID string, userID int, metadata map[string]interface{}) {
	// Create a span for the business event
	ctx, span := p.StartRepositorySpan(ctx, fmt.Sprintf("event.%s", event), entity, []attribute.KeyValue{
		attribute.String("event", event),
		attribute.String("entity", entity),
		attribute.String("entity_id", entityID),
		attribute.Int("user_id", userID),
	})

	// Add metadata as span attributes
	for key, value := range metadata {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(fmt.Sprintf("meta.%s", key), v))
		case int:
			span.SetAttributes(attribute.Int(fmt.Sprintf("meta.%s", key), v))
		case bool:
			span.SetAttributes(attribute.Bool(fmt.Sprintf("meta.%s", key), v))
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

func (p *OTELProbe) RecordHTTPOperation(ctx context.Context, method string, path string, statusCode int, duration time.Duration) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.duration_ns", duration.Nanoseconds()),
	)

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

func (p *OTELProbe) RecordError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	p.logger.ErrorContext(ctx, "Operation error recorded",
		"operation", operation,
		"error", err,
		"metadata", metadata)
}
