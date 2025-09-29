package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"todoapp/internal/adapter/http"
	"todoapp/internal/adapter/telemetry"
	"todoapp/pkg/config"
)

func main() {
	ctx := context.Background()

	logger, err := config.NewLokiLogger("todoapp", "http://localhost:3100")

	if err != nil {
		log.Fatal("Failed to initialize Loki logger:", err)
	}

	defer logger.Sync()

	telemetryContainer, err := telemetry.NewContainer(telemetry.Config{
		ServiceName:    "todoapp",
		ServiceVersion: "1.0.0",
		MetricsPort:    "9091",
		OTLPEndpoint:   "localhost:4317",
	}, slog.Default())

	if err != nil {
		log.Fatal("Failed to initialize telemetry:", err)
	}

	defer telemetryContainer.Shutdown(ctx)

	telemetryContainer.AppMetrics.StartSystemMetrics(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		config := config.GetDefaultConfig()

		if os.Getenv("GIN_MODE") == "release" {
			config.Environment = "production"
			config.EnforceHTTPS = true
		}

		http.StartServerWithConfig(telemetryContainer.AppMetrics, logger, config)
	}()

	<-c
	logger.Logger.Info("Shutting down gracefully...")
}
