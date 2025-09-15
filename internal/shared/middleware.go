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
	// OpenTelemetry tracing middleware
	router.Use(otelgin.Middleware(serviceName))

	// Logging middleware
	router.Use(LoggingMiddleware(logger))

	// Custom metrics middleware
	router.Use(MetricsMiddleware(metrics))
}
