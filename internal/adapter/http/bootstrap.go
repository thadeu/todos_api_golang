package http

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	database "todoapp/internal/adapter/database/sqlite"
	"todoapp/internal/adapter/http/routes"

	"todoapp/internal/core/telemetry"
	"todoapp/pkg/config"
)

func StartServer(metrics *telemetry.AppMetrics, logger *config.LokiLogger) {
	StartServerWithConfig(metrics, logger, config.GetDefaultConfig())
}

func StartServerWithConfig(metrics *telemetry.AppMetrics, logger *config.LokiLogger, config *config.AppConfig) {
	db, _ := database.NewDB()
	defer db.Close()

	container := NewContainer(db, logger)

	router := routes.SetupRouterWithConfig(routes.HandlersConfig{
		AuthHandler: container.AuthHandler,
		TodoHandler: container.TodoHandler,
	}, metrics, logger, config)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	slog.Info("Server starting",
		"port", port,
		"environment", config.Environment,
		"rate_limit_enabled", config.RateLimitEnabled,
		"https_enforced", config.EnforceHTTPS)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err)
	}
}
