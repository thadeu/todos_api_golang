package config

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	. "todoapp/pkg"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type RateLimitEndpointConfig struct {
	Requests int
	Window   time.Duration
	KeyFunc  func(*gin.Context) string
}

type RateLimiter struct {
	cache   *cache.Cache
	config  map[string]RateLimitEndpointConfig
	logger  *zap.Logger
	metrics *AppMetrics
	mutex   sync.RWMutex
}

type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

func NewRateLimiter(logger *zap.Logger, metrics *AppMetrics) *RateLimiter {
	c := cache.New(5*time.Minute, 10*time.Minute)

	configs := map[string]RateLimitEndpointConfig{
		"POST /signup": {
			Requests: 5,
			Window:   time.Minute,
			KeyFunc:  GetClientIP,
		},
		"POST /auth": {
			Requests: 10,
			Window:   time.Minute,
			KeyFunc:  GetClientIP,
		},
		"GET /todos": {
			Requests: 100,
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"POST /todos": {
			Requests: 20,
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"PUT /todo/:uuid": {
			Requests: 10,
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"DELETE /todos/:uuid": {
			Requests: 5,
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"/todos": {
			Requests: 100,
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"default": {
			Requests: 60,
			Window:   time.Minute,
			KeyFunc:  GetClientIP,
		},
	}

	return &RateLimiter{
		cache:   c,
		config:  configs,
		logger:  logger,
		metrics: metrics,
		mutex:   sync.RWMutex{},
	}
}

func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		normalizedPath := rl.normalizePath(path)
		methodPath := c.Request.Method + " " + normalizedPath

		config, exists := rl.config[methodPath]
		if !exists {
			config, exists = rl.config[normalizedPath]
			if !exists {
				config = rl.config["default"]
			}
		}

		key := rl.generateKey(c, methodPath, config.KeyFunc)

		rl.logger.Info("Rate limit check",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("normalizedPath", normalizedPath),
			zap.String("methodPath", methodPath),
			zap.String("key", key),
			zap.Int("limit", config.Requests),
			zap.Duration("window", config.Window))

		allowed, remaining, resetTime, err := rl.checkRateLimit(key, config)
		if err != nil {
			rl.logger.Error("Rate limit check failed",
				zap.String("key", key),
				zap.String("path", path),
				zap.Error(err))
			c.Next()
			return
		}

		keyType := "ip"
		if strings.Contains(key, "user_") {
			keyType = "user"
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			if rl.metrics != nil {
				rl.metrics.RecordRateLimitHit(c.Request.Context(), path, keyType)
			}

			rl.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.String("path", path),
				zap.Int("limit", config.Requests),
				zap.Duration("window", config.Window))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", config.Requests, config.Window),
				"retry_after": int(time.Until(resetTime).Seconds()),
			})
			c.Abort()
			return
		}

		if rl.metrics != nil {
			rl.metrics.RecordRateLimitAllowed(c.Request.Context(), path, keyType)
		}

		c.Next()
	}
}

func (rl *RateLimiter) checkRateLimit(key string, config RateLimitEndpointConfig) (bool, int, time.Time, error) {
	now := time.Now()

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if entry, found := rl.cache.Get(key); found {
		rateLimitEntry := entry.(RateLimitEntry)

		if now.After(rateLimitEntry.ResetTime) {

			resetTime := now.Add(config.Window)
			newEntry := RateLimitEntry{
				Count:     1,
				ResetTime: resetTime,
			}
			rl.cache.Set(key, newEntry, config.Window)
			return true, config.Requests - 1, resetTime, nil
		}

		if rateLimitEntry.Count >= config.Requests {
			return false, 0, rateLimitEntry.ResetTime, nil
		}

		rateLimitEntry.Count++
		rl.cache.Set(key, rateLimitEntry, cache.DefaultExpiration)

		return true, config.Requests - rateLimitEntry.Count, rateLimitEntry.ResetTime, nil
	}

	resetTime := now.Add(config.Window)
	newEntry := RateLimitEntry{
		Count:     1,
		ResetTime: resetTime,
	}
	rl.cache.Set(key, newEntry, config.Window)

	return true, config.Requests - 1, resetTime, nil
}

func (rl *RateLimiter) normalizePath(path string) string {

	if strings.HasPrefix(path, "/todo/") {
		// /todo/123 -> /todo/:uuid
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			parts[2] = ":uuid"
			return strings.Join(parts, "/")
		}
	}
	if strings.HasPrefix(path, "/todos/") {
		// /todos/123 -> /todos/:uuid
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			parts[2] = ":uuid"
			return strings.Join(parts, "/")
		}
	}
	return path
}

func (rl *RateLimiter) generateKey(c *gin.Context, path string, keyFunc func(*gin.Context) string) string {
	identifier := keyFunc(c)
	return fmt.Sprintf("rate_limit:%s:%s", path, identifier)
}

func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("x-user-id"); exists {
		return fmt.Sprintf("user_%v", userID)
	}
	return GetClientIP(c)
}

func (rl *RateLimiter) SetConfig(path string, config RateLimitEndpointConfig) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.config[path] = config
}

func (rl *RateLimiter) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	activeEntries := rl.cache.ItemCount()

	stats["active_entries"] = activeEntries
	stats["configs"] = len(rl.config)

	return stats
}
