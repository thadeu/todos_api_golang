package telemetry

import (
	"context"
	"log/slog"
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

	"todos/internal/core/port"
	"todos/internal/core/telemetry"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	MetricsPort    string
	OTLPEndpoint   string
}

type Container struct {
	TracerProvider     *sdktrace.TracerProvider
	MeterProvider      *sdkmetric.MeterProvider
	PrometheusRegistry *prometheus.Registry
	MetricsServer      *http.Server
	AppMetrics         *telemetry.AppMetrics
}

func NewContainer(config Config, logger *slog.Logger) (*Container, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
		semconv.DeploymentEnvironmentKey.String("development"),
	)

	registry := prometheus.NewRegistry()
	appMetrics := telemetry.NewAppMetrics(registry)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	ctx := context.Background()

	otlpExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otlpExporter,
			sdktrace.WithBatchTimeout(1*time.Second),
			sdktrace.WithMaxExportBatchSize(1),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tracerProvider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))

	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	metricsServer := &http.Server{
		Addr:         ":" + config.MetricsPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start metrics server", "error", err)
		}
	}()

	return &Container{
		TracerProvider:     tracerProvider,
		MeterProvider:      meterProvider,
		PrometheusRegistry: registry,
		MetricsServer:      metricsServer,
		AppMetrics:         appMetrics,
	}, nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	if err := c.TracerProvider.Shutdown(ctx); err != nil {
		return err
	}

	if err := c.MeterProvider.Shutdown(ctx); err != nil {
		return err
	}

	if err := c.MetricsServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// NewTelemetryProbe cria um probe de telemetria configurado
func (c *Container) NewTelemetryProbe(logger *slog.Logger) port.Telemetry {
	return telemetry.NewOTELProbe(logger)
}
