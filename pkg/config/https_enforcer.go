package config

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HTTPSEnforcer struct {
	enabled bool
	logger  *zap.Logger
}

func NewHTTPSEnforcer(logger *zap.Logger) *HTTPSEnforcer {

	env := os.Getenv("GIN_MODE")
	enabled := env == "release" || os.Getenv("ENFORCE_HTTPS") == "true"

	return &HTTPSEnforcer{
		enabled: enabled,
		logger:  logger,
	}
}

func (he *HTTPSEnforcer) HTTPSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !he.enabled {
			c.Next()
			return
		}

		if c.Request.TLS != nil {
			c.Next()
			return
		}

		proto := c.GetHeader("X-Forwarded-Proto")
		if proto == "https" {
			c.Next()
			return
		}

		host := c.GetHeader("Host")
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			c.Next()
			return
		}

		httpsURL := "https://" + host + c.Request.RequestURI

		he.logger.Info("Redirecting to HTTPS",
			zap.String("original_url", c.Request.URL.String()),
			zap.String("https_url", httpsURL),
			zap.String("user_agent", c.GetHeader("User-Agent")))

		c.Redirect(http.StatusMovedPermanently, httpsURL)
		c.Abort()
	}
}

func (he *HTTPSEnforcer) SetEnabled(enabled bool) {
	he.enabled = enabled
}

func (he *HTTPSEnforcer) IsEnabled() bool {
	return he.enabled
}
