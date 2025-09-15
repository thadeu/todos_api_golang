package shared

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// RateLimitEndpointConfig configuration for rate limiting per endpoint
type RateLimitEndpointConfig struct {
	Requests int
	Window   time.Duration
	KeyFunc  func(*gin.Context) string
}

// RateLimiter structure for managing rate limiting
type RateLimiter struct {
	cache   *cache.Cache
	config  map[string]RateLimitEndpointConfig
	logger  *zap.Logger
	metrics *AppMetrics
	mutex   sync.RWMutex
}

// RateLimitEntry cache entry for rate limiting
type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

// NewRateLimiter creates a new rate limiter instance
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

// RateLimitMiddleware middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Normalize path for pattern matching
		normalizedPath := rl.normalizePath(path)
		methodPath := c.Request.Method + " " + normalizedPath

		// Find configuration
		config, exists := rl.config[methodPath]
		if !exists {
			config, exists = rl.config[normalizedPath]
			if !exists {
				config = rl.config["default"]
			}
		}

		// Generate key for rate limiting
		key := rl.generateKey(c, methodPath, config.KeyFunc)

		// Debug logging for troubleshooting
		rl.logger.Info("Rate limit check",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("normalizedPath", normalizedPath),
			zap.String("methodPath", methodPath),
			zap.String("key", key),
			zap.Int("limit", config.Requests),
			zap.Duration("window", config.Window))

		// Check rate limit
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

		// Add informative headers
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

// checkRateLimit checks if request is within limit
func (rl *RateLimiter) checkRateLimit(key string, config RateLimitEndpointConfig) (bool, int, time.Time, error) {
	now := time.Now()

	// Use write lock to prevent race conditions
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Get existing entry
	if entry, found := rl.cache.Get(key); found {
		rateLimitEntry := entry.(RateLimitEntry)

		// Check if window has expired
		if now.After(rateLimitEntry.ResetTime) {
			// Window expired, create new entry
			resetTime := now.Add(config.Window)
			newEntry := RateLimitEntry{
				Count:     1,
				ResetTime: resetTime,
			}
			rl.cache.Set(key, newEntry, config.Window)
			return true, config.Requests - 1, resetTime, nil
		}

		// Check if limit exceeded
		if rateLimitEntry.Count >= config.Requests {
			return false, 0, rateLimitEntry.ResetTime, nil
		}

		// Increment counter
		rateLimitEntry.Count++
		rl.cache.Set(key, rateLimitEntry, cache.DefaultExpiration)

		return true, config.Requests - rateLimitEntry.Count, rateLimitEntry.ResetTime, nil
	}

	// Create new entry
	resetTime := now.Add(config.Window)
	newEntry := RateLimitEntry{
		Count:     1,
		ResetTime: resetTime,
	}
	rl.cache.Set(key, newEntry, config.Window)

	return true, config.Requests - 1, resetTime, nil
}

// normalizePath normalizes a path by replacing UUIDs with :uuid pattern
func (rl *RateLimiter) normalizePath(path string) string {
	// Handle specific patterns
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

// generateKey generates unique key for rate limiting
func (rl *RateLimiter) generateKey(c *gin.Context, path string, keyFunc func(*gin.Context) string) string {
	identifier := keyFunc(c)
	return fmt.Sprintf("rate_limit:%s:%s", path, identifier)
}

// getUserID extracts authenticated user ID
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("x-user-id"); exists {
		return fmt.Sprintf("user_%v", userID)
	}
	return GetClientIP(c)
}

// SetConfig allows configuring rate limits for specific endpoints
func (rl *RateLimiter) SetConfig(path string, config RateLimitEndpointConfig) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.config[path] = config
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count active entries in cache
	activeEntries := rl.cache.ItemCount()

	stats["active_entries"] = activeEntries
	stats["configs"] = len(rl.config)

	return stats
}
