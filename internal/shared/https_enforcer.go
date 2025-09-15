package shared

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPSEnforcer middleware para forçar HTTPS em produção
type HTTPSEnforcer struct {
	enabled bool
	logger  *zap.Logger
}

// NewHTTPSEnforcer cria uma nova instância do HTTPS enforcer
func NewHTTPSEnforcer(logger *zap.Logger) *HTTPSEnforcer {
	// Verificar se está em produção
	env := os.Getenv("GIN_MODE")
	enabled := env == "release" || os.Getenv("ENFORCE_HTTPS") == "true"

	return &HTTPSEnforcer{
		enabled: enabled,
		logger:  logger,
	}
}

// HTTPSMiddleware middleware para forçar HTTPS
func (he *HTTPSEnforcer) HTTPSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Se não estiver habilitado, continuar normalmente
		if !he.enabled {
			c.Next()
			return
		}

		// Verificar se a requisição já é HTTPS
		if c.Request.TLS != nil {
			c.Next()
			return
		}

		// Verificar headers de proxy (para load balancers)
		proto := c.GetHeader("X-Forwarded-Proto")
		if proto == "https" {
			c.Next()
			return
		}

		// Verificar se é localhost (desenvolvimento)
		host := c.GetHeader("Host")
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			c.Next()
			return
		}

		// Redirecionar para HTTPS
		httpsURL := "https://" + host + c.Request.RequestURI

		he.logger.Info("Redirecting to HTTPS",
			zap.String("original_url", c.Request.URL.String()),
			zap.String("https_url", httpsURL),
			zap.String("user_agent", c.GetHeader("User-Agent")))

		c.Redirect(http.StatusMovedPermanently, httpsURL)
		c.Abort()
	}
}

// SetEnabled permite habilitar/desabilitar o enforcement
func (he *HTTPSEnforcer) SetEnabled(enabled bool) {
	he.enabled = enabled
}

// IsEnabled retorna se o enforcement está habilitado
func (he *HTTPSEnforcer) IsEnabled() bool {
	return he.enabled
}
