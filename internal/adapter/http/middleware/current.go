package middleware

import (
	ct "todos/pkg/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CurrentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		current := ct.NewCurrent()

		current.Set("request_id", c.GetHeader("X-Request-ID"))

		if requestID := current.Get("request_id"); requestID == "" {
			current.Set("request_id", uuid.New().String())
		}

		current.Set("user_agent", c.Request.UserAgent())
		current.Set("ip_address", c.ClientIP())
		current.Set("method", c.Request.Method)
		current.Set("path", c.Request.URL.Path)

		ct.SetGlobalCurrent(current)

		ctx := ct.SetCurrent(c.Request.Context(), current)
		c.Request = c.Request.WithContext(ctx)

		c.Set("current", current)

		c.Next()

		ct.ResetGlobalCurrent()
	}
}

func GetCurrent(c *gin.Context) *ct.Current {
	if current, ok := c.Get("current"); ok {
		if curr, ok := current.(*ct.Current); ok {
			return curr
		}
	}

	if current, ok := ct.FromContext(c.Request.Context()); ok {
		return current
	}

	return ct.NewCurrent()
}
