package api

import (
	"log/slog"
	"net/http"
	"time"

	"todoapp/internal/delivery/http/routes"
	"todoapp/internal/infrastructure"
	. "todoapp/pkg"
	. "todoapp/pkg/config"
	. "todoapp/pkg/db"
	. "todoapp/pkg/tracing"
)

func StartServer(metrics *AppMetrics, logger *LokiLogger) {
	StartServerWithConfig(metrics, logger, GetDefaultConfig())
}

func StartServerWithConfig(metrics *AppMetrics, logger *LokiLogger, config *AppConfig) {
	db := InitDB()

	// Initialize dependency container with Clean Architecture
	container := infrastructure.NewContainer(db, logger)

	router := routes.SetupRouterWithConfig(routes.HandlersConfig{
		AuthHandler: container.AuthHandler,
		TodoHandler: container.TodoHandler,
	}, metrics, logger, config)

	port := GetServerPort()

	slog.Info("Server starting",
		"port", port,
		"environment", config.Environment,
		"rate_limit_enabled", config.RateLimitEnabled,
		"cache_enabled", config.CacheEnabled,
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
