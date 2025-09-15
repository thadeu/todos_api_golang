package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	MetricsPort    string
	OTLPEndpoint   string
}

type Telemetry struct {
	TracerProvider     *sdktrace.TracerProvider
	MeterProvider      *sdkmetric.MeterProvider
	PrometheusRegistry *prometheus.Registry
	Server             *http.Server
}

func InitTelemetry(config TelemetryConfig) (*Telemetry, error) {
	// Resource identifica o serviço
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
		semconv.DeploymentEnvironmentKey.String("development"),
	)

	// Setup Prometheus registry
	registry := prometheus.NewRegistry()

	// Setup meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// Setup OTLP gRPC exporter for Tempo
	ctx := context.Background()
	otlpExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Setup tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otlpExporter,
			sdktrace.WithBatchTimeout(1*time.Second),
			sdktrace.WithMaxExportBatchSize(1), // Force flush after each span
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tracerProvider)

	// Inicializar métricas de runtime
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		return nil, err
	}

	// Setup HTTP server para métricas do Prometheus
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:         ":" + config.MetricsPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Iniciar servidor de métricas em goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't fail the application
			// TODO: usar logger adequado
		}
	}()

	return &Telemetry{
		TracerProvider:     tracerProvider,
		MeterProvider:      meterProvider,
		PrometheusRegistry: registry,
		Server:             server,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	// Shutdown tracer provider
	if err := t.TracerProvider.Shutdown(ctx); err != nil {
		return err
	}

	// Shutdown meter provider
	if err := t.MeterProvider.Shutdown(ctx); err != nil {
		return err
	}

	// Shutdown HTTP server
	if err := t.Server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
