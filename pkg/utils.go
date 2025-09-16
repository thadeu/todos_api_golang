package pkg

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetServerPort() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return port
}

func GetClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {

		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	ip := c.ClientIP()

	if ip == "" {
		return "unknown"
	}

	return ip
}
