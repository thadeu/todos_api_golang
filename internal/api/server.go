package api

import (
	"log/slog"
	"net/http"
	"time"

	. "todoapp/internal"
	. "todoapp/internal/handlers"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/shared"
)

func StartServer(metrics *AppMetrics, logger *LokiLogger) {
	StartServerWithConfig(metrics, logger, GetDefaultConfig())
}

func StartServerWithConfig(metrics *AppMetrics, logger *LokiLogger, config *AppConfig) {
	db := InitDB()

	user := NewUserRepository(db)
	todo := NewTodoRepository(db)

	todoService := NewTodoService(todo)
	authService := NewAuthService(user)

	todoHandler := NewTodoHandler(todoService, logger)
	authHandler := NewAuthHandler(authService)

	router := SetupRouterWithConfig(HandlersConfig{
		AuthHandler: authHandler,
		TodoHandler: todoHandler,
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
