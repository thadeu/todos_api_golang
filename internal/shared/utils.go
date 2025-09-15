package shared

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP extrai IP do cliente considerando headers de proxy
func GetClientIP(c *gin.Context) string {
	// Verificar headers de proxy primeiro
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// Pegar o primeiro IP da lista
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback para RemoteAddr
	ip := c.ClientIP()
	if ip == "" {
		return "unknown"
	}

	return ip
}
