package shared

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func MetricsMiddleware(metrics *AppMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		metrics.IncrementActiveConnections(c.Request.Context())
		defer metrics.DecrementActiveConnections(c.Request.Context())

		// Process request
		c.Next()

		// Record metrics
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
	// HTTPS Enforcement (deve ser o primeiro)
	httpsEnforcer := NewHTTPSEnforcer(logger.Logger.Logger)
	router.Use(httpsEnforcer.HTTPSMiddleware())

	// OpenTelemetry tracing middleware
	router.Use(otelgin.Middleware(serviceName))

	// Logging middleware
	router.Use(LoggingMiddleware(logger))

	// Rate Limiting middleware
	if config.RateLimitEnabled {
		rateLimiter := NewRateLimiter(logger.Logger.Logger, metrics)
		router.Use(rateLimiter.RateLimitMiddleware())
	}

	// Response Cache middleware (após rate limiting, antes de autenticação)
	if config.CacheEnabled {
		responseCache := NewResponseCache(logger.Logger.Logger, metrics)
		// Aplicar configurações específicas
		for path, cacheConfig := range config.CacheConfigs {
			responseCache.SetConfig(path, cacheConfig)
		}
		router.Use(responseCache.CacheMiddleware())
	}

	// Custom metrics middleware
	router.Use(MetricsMiddleware(metrics))
}
