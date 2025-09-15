package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LokiLogger struct {
	Logger      *otelzap.Logger
	serviceName string
	lokiURL     string
	httpClient  *http.Client
}

type LokiLogEntry struct {
	Streams []LokiStream `json:"streams"`
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func NewLokiLogger(serviceName, lokiURL string) (*LokiLogger, error) {
	// Configuração do Zap
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	// Wrapper com otelzap para correlação automática de traces
	otelLogger := otelzap.New(zapLogger)

	return &LokiLogger{
		Logger:      otelLogger,
		serviceName: serviceName,
		lokiURL:     lokiURL + "/loki/api/v1/push",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (l *LokiLogger) Sync() error {
	return l.Logger.Sync()
}

func (l *LokiLogger) InfoWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.logWithTrace(ctx, zapcore.InfoLevel, msg, fields...)
}

func (l *LokiLogger) ErrorWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.logWithTrace(ctx, zapcore.ErrorLevel, msg, fields...)
}

func (l *LokiLogger) logWithTrace(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	logFields := append(fields,
		zap.String("service", l.serviceName),
		zap.String("level", level.String()),
	)

	// Log localmente com otelzap (inclui automaticamente trace_id e span_id)
	switch level {
	case zapcore.InfoLevel:
		l.Logger.Ctx(ctx).Info(msg, logFields...)
	case zapcore.ErrorLevel:
		l.Logger.Ctx(ctx).Error(msg, logFields...)
	default:
		l.Logger.Ctx(ctx).Info(msg, logFields...)
	}

	// Enviar para Loki (simplificado)
	go l.sendToLokiSimple(ctx, level, msg, logFields)
}

func (l *LokiLogger) sendToLoki(ctx context.Context, level zapcore.Level, msg string, fields []zap.Field) {
	// Criar um objeto de log simples
	logData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"level":     level.String(),
		"message":   msg,
		"service":   l.serviceName,
	}

	// Adicionar trace context se disponível
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logData["trace_id"] = span.SpanContext().TraceID().String()
		logData["span_id"] = span.SpanContext().SpanID().String()
	}

	// Adicionar campos extras
	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			logData[field.Key] = field.String
		case zapcore.Int64Type:
			logData[field.Key] = field.Integer
		case zapcore.BoolType:
			logData[field.Key] = field.Integer == 1
		case zapcore.Float64Type:
			logData[field.Key] = field.Interface
		case zapcore.DurationType:
			logData[field.Key] = field.Interface
		case zapcore.ErrorType:
			logData[field.Key] = field.Interface
		default:
			logData[field.Key] = fmt.Sprintf("%v", field.Interface)
		}
	}

	// Serializar para JSON
	jsonBytes, err := json.Marshal(logData)
	if err != nil {
		l.Logger.Ctx(ctx).Error("Failed to marshal log data", zap.Error(err))
		return
	}

	// Loki expects a specific JSON format
	lokiEntry := LokiLogEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"service": l.serviceName,
					"level":   level.String(),
				},
				Values: [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), string(jsonBytes)},
				},
			},
		},
	}

	l.sendToLokiHTTP(lokiEntry)
}

func (l *LokiLogger) sendToLokiSimple(ctx context.Context, level zapcore.Level, msg string, fields []zap.Field) {
	fmt.Printf("DEBUG: sendToLokiSimple called with msg: %s\n", msg)

	// Criar um log simples para Loki
	logLine := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s","service":"%s"`,
		time.Now().Format(time.RFC3339Nano),
		level.String(),
		msg,
		l.serviceName)

	// Adicionar trace context se disponível
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logLine += fmt.Sprintf(`,"trace_id":"%s","span_id":"%s"`,
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String())
	}

	// Adicionar campos extras
	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			logLine += fmt.Sprintf(`,"%s":"%s"`, field.Key, field.String)
		case zapcore.Int64Type:
			logLine += fmt.Sprintf(`,"%s":%d`, field.Key, field.Integer)
		case zapcore.BoolType:
			logLine += fmt.Sprintf(`,"%s":%t`, field.Key, field.Integer == 1)
		case zapcore.DurationType:
			if field.Interface != nil {
				logLine += fmt.Sprintf(`,"%s":"%v"`, field.Key, field.Interface)
			} else {
				logLine += fmt.Sprintf(`,"%s":"0s"`, field.Key)
			}
		default:
			logLine += fmt.Sprintf(`,"%s":"%v"`, field.Key, field.Interface)
		}
	}

	logLine += "}"

	// Criar entrada para Loki
	lokiEntry := LokiLogEntry{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"service": l.serviceName,
					"level":   level.String(),
				},
				Values: [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), logLine},
				},
			},
		},
	}

	fmt.Printf("DEBUG: About to send to Loki: %+v\n", lokiEntry)
	l.sendToLokiHTTP(lokiEntry)
}

func (l *LokiLogger) sendToLokiHTTP(lokiEntry LokiLogEntry) {
	// Serializar para JSON
	body, err := json.Marshal(lokiEntry)
	if err != nil {
		return
	}

	// Enviar para Loki
	req, err := http.NewRequest("POST", l.lokiURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Consumir resposta se necessário
	io.ReadAll(resp.Body)
}

// Helper functions para compatibilidade
func LogError(ctx context.Context, logger *LokiLogger, err error, msg string, fields ...zap.Field) {
	logger.ErrorWithTrace(ctx, msg, append(fields, zap.Error(err))...)
}

func LogInfo(ctx context.Context, logger *LokiLogger, msg string, fields ...zap.Field) {
	logger.InfoWithTrace(ctx, msg, fields...)
}
