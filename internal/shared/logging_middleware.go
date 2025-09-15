package shared

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingMiddleware cria um middleware para logs estruturados
func LoggingMiddleware(logger *LokiLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Processar a requisição
		c.Next()

		// Calcular duração
		latency := time.Since(start)

		// Construir query string se existir
		if raw != "" {
			path = path + "?" + raw
		}

		// Log estruturado com trace context usando otelzap
		logger.Logger.Ctx(c.Request.Context()).Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("service", logger.serviceName),
		)

		// Enviar para Loki também
		go logger.sendToLokiSimple(c.Request.Context(), zapcore.InfoLevel, "HTTP Request",
			[]zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("service", logger.serviceName),
			})
	}
}
