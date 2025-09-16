package middlewares

import (
	"time"

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
