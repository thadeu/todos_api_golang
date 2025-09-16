package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	api "todoapp/pkg/api"
	. "todoapp/pkg/config"
	. "todoapp/pkg/tracing"
)

func main() {
	ctx := context.Background()

	logger, err := NewLokiLogger("todoapp", "http://localhost:3100")

	if err != nil {
		log.Fatal("Failed to initialize Loki logger:", err)
	}

	defer logger.Sync()

	telemetry, err := InitTelemetry(TelemetryConfig{
		ServiceName:    "todoapp",
		ServiceVersion: "1.0.0",
		MetricsPort:    "9091",
		OTLPEndpoint:   "localhost:4317",
	})

	if err != nil {
		log.Fatal("Failed to initialize telemetry:", err)
	}

	defer telemetry.Shutdown(ctx)

	metrics := NewAppMetrics(telemetry.PrometheusRegistry)
	metrics.StartSystemMetrics(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		config := GetDefaultConfig()

		if os.Getenv("GIN_MODE") == "release" {
			config.Environment = "production"
			config.EnforceHTTPS = true
		}

		api.StartServerWithConfig(metrics, logger, config)
	}()

	<-c
	logger.Logger.Info("Shutting down gracefully...")
}
