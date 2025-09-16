package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerHelper helper para facilitar criação de spans
type TracerHelper struct {
	tracer trace.Tracer
}

// NewTracerHelper cria uma nova instância do helper
func NewTracerHelper(serviceName string) *TracerHelper {
	return &TracerHelper{
		tracer: otel.Tracer(serviceName),
	}
}

// StartSpan cria um novo span com contexto
func (th *TracerHelper) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return th.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes cria um span com atributos iniciais
func (th *TracerHelper) StartSpanWithAttributes(ctx context.Context, name string, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(opts, trace.WithAttributes(attrs...))
	return th.tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes adiciona atributos a um span
func AddSpanAttributes(span trace.Span, attrs []attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanError marca um span como erro
func AddSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// AddSpanEvent adiciona um evento ao span
func AddSpanEvent(span trace.Span, name string, attrs []attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// AddBusinessAttributes adiciona atributos de negócio comuns
func AddBusinessAttributes(span trace.Span, userID int, operation string) {
	attrs := []attribute.KeyValue{
		attribute.Int("user.id", userID),
		attribute.String("operation", operation),
	}
	span.SetAttributes(attrs...)
}

// AddDatabaseAttributes adiciona atributos específicos de banco de dados
func AddDatabaseAttributes(span trace.Span, table string, operation string, query string) {
	attrs := []attribute.KeyValue{
		attribute.String("db.table", table),
		attribute.String("db.operation", operation),
		attribute.String("db.query", query),
	}
	span.SetAttributes(attrs...)
}

// AddHTTPAttributes adiciona atributos específicos de HTTP
func AddHTTPAttributes(span trace.Span, method string, url string, statusCode int) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.url", url),
		attribute.Int("http.status_code", statusCode),
	}
	span.SetAttributes(attrs...)
}

// GetTraceID extrai trace ID do contexto
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extrai span ID do contexto
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// CreateChildSpan cria um span filho com contexto
func CreateChildSpan(ctx context.Context, name string, attrs []attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("todoapp")
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
	}
	return tracer.Start(ctx, name, opts...)
}

// SpanWrapper wrapper para executar funções com span
func SpanWrapper(ctx context.Context, name string, attrs []attribute.KeyValue, fn func(context.Context) error) error {
	ctx, span := CreateChildSpan(ctx, name, attrs)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		AddSpanError(span, err)
	}

	return err
}

// DatabaseSpanWrapper wrapper específico para operações de banco
func DatabaseSpanWrapper(ctx context.Context, table, operation, query string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("db.table", table),
		attribute.String("db.operation", operation),
		attribute.String("db.query", query),
		attribute.String("db.system", "sqlite"),
	}

	return SpanWrapper(ctx, fmt.Sprintf("db.%s.%s", table, operation), attrs, fn)
}

// ServiceSpanWrapper wrapper específico para operações de serviço
func ServiceSpanWrapper(ctx context.Context, service, operation string, userID int, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("service.operation", operation),
		attribute.Int("user.id", userID),
	}

	return SpanWrapper(ctx, fmt.Sprintf("service.%s.%s", service, operation), attrs, fn)
}
