package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	. "todoapp/pkg/auth"

	. "todoapp/pkg/config"
	. "todoapp/pkg/response"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func MetricsMiddleware(metrics *AppMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		metrics.IncrementActiveConnections(c.Request.Context())
		defer metrics.DecrementActiveConnections(c.Request.Context())

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		metrics.RecordRequest(
			c.Request.Context(),
			c.Request.Method,
			c.FullPath(),
			string(rune(status)),
			duration,
		)
	}
}

func SetupGinMiddleware(router *gin.Engine, serviceName string, metrics *AppMetrics, logger *LokiLogger) {
	SetupGinMiddlewareWithConfig(router, serviceName, metrics, logger, GetDefaultConfig())
}

func SetupGinMiddlewareWithConfig(router *gin.Engine, serviceName string, metrics *AppMetrics, logger *LokiLogger, config *AppConfig) {

	httpsEnforcer := NewHTTPSEnforcer(logger.Logger.Logger)
	router.Use(httpsEnforcer.HTTPSMiddleware())

	router.Use(otelgin.Middleware(serviceName))

	router.Use(LoggingMiddleware(logger))

	if config.CacheEnabled {
		responseCache := NewResponseCache(logger.Logger.Logger, metrics)
		for path, cacheConfig := range config.CacheConfigs {
			responseCache.SetConfig(path, cacheConfig)
		}
		router.Use(responseCache.CacheMiddleware())
	}

	if config.RateLimitEnabled {
		rateLimiter := NewRateLimiter(logger.Logger.Logger, metrics)
		router.Use(rateLimiter.RateLimitMiddleware())
	}

	router.Use(MetricsMiddleware(metrics))
}

func JwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")

		if bearer == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Unauthorized request"}})
			return
		}

		token, err := VerifyJwtToken(bearer[len("Bearer "):])

		if err != nil {
			slog.Info("Error", "error", err)

			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Unauthorized request", err.Error()}})
			return
		}

		userId := int(token["user_id"].(float64))
		context := context.WithValue(r.Context(), "x-user-id", userId)

		next.ServeHTTP(w, r.WithContext(context))
	}
}

func GinJwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")

		if bearer == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Unauthorized request"},
			})

			c.Abort()
			return
		}

		if !strings.HasPrefix(bearer, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Invalid authorization format"},
			})

			c.Abort()
			return
		}

		token, err := VerifyJwtToken(bearer[len("Bearer "):])

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Unauthorized request", err.Error()},
			})
			c.Abort()
			return
		}

		userId := int(token["user_id"].(float64))

		c.Set("x-user-id", userId)
		c.Next()
	}
}
