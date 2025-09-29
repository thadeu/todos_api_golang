package config

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"todos/internal/core/telemetry"
	. "todos/pkg"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestNewRateLimiter(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	Expect(rl).ToNot(BeNil())
	Expect(rl.cache).ToNot(BeNil())
	Expect(rl.config).ToNot(BeNil())
	Expect(rl.logger).ToNot(BeNil())
	Expect(rl.metrics).ToNot(BeNil())
}

func TestRateLimitMiddleware_AllowedRequests(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Test allowed requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(w.Header().Get("X-RateLimit-Limit")).ToNot(BeEmpty())
		Expect(w.Header().Get("X-RateLimit-Remaining")).ToNot(BeEmpty())
	}
}

func TestRateLimitMiddleware_ExceedLimit(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Exceed rate limit (default is 60 requests per minute)
	for i := 0; i < 65; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if i < 60 {
			Expect(w.Code).To(Equal(200))
		} else {
			Expect(w.Code).To(Equal(429))
		}
	}
}

func TestRateLimitMiddleware_UserBasedLimiting(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("x-user-id", "123")
		c.Next()
	})
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.POST("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(201, gin.H{"method": "POST", "count": callCount})
	})

	// Test POST requests - should use user-based rate limiting (20 requests per minute)
	expectedRemaining := []int{19, 18, 17, 16, 15}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(`{"title":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(201))
		Expect(callCount).To(Equal(i + 1))

		remaining := w.Header().Get("X-RateLimit-Remaining")
		expectedRemainingStr := strconv.Itoa(expectedRemaining[i])
		Expect(remaining).To(Equal(expectedRemainingStr),
			"POST Request %d: Expected remaining %s, got %s",
			i+1, expectedRemainingStr, remaining)
	}
}

func TestRateLimitMiddleware_PUTRequests(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("x-user-id", "123")
		c.Next()
	})
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.PUT("/todo/:uuid", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"method": "PUT", "count": callCount})
	})

	// Test PUT requests - should use limit 10 per minute
	expectedRemaining := []int{9, 8, 7, 6, 5}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/todo/123", strings.NewReader(`{"title":"updated"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(callCount).To(Equal(i + 1))

		remaining := w.Header().Get("X-RateLimit-Remaining")
		expectedRemainingStr := strconv.Itoa(expectedRemaining[i])
		Expect(remaining).To(Equal(expectedRemainingStr),
			"PUT Request %d: Expected remaining %s, got %s",
			i+1, expectedRemainingStr, remaining)
	}
}

func TestRateLimitMiddleware_DELETERequests(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("x-user-id", "123")
		c.Next()
	})
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.DELETE("/todos/:uuid", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"method": "DELETE", "count": callCount})
	})

	// Test DELETE requests - should use limit 5 per minute
	expectedRemaining := []int{4, 3, 2, 1, 0}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/todos/123", nil)
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(callCount).To(Equal(i + 1))

		remaining := w.Header().Get("X-RateLimit-Remaining")
		expectedRemainingStr := strconv.Itoa(expectedRemaining[i])
		Expect(remaining).To(Equal(expectedRemainingStr),
			"DELETE Request %d: Expected remaining %s, got %s",
			i+1, expectedRemainingStr, remaining)
	}
}

func TestRateLimitMiddleware_WindowReset(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make requests to consume rate limit
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
	}

	// Wait for window to reset (using a short window for testing)
	time.Sleep(100 * time.Millisecond)

	// After reset, should be able to make requests again
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	Expect(w.Code).To(Equal(200))
}

func TestRateLimiterGetStats(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	stats := rl.GetStats()
	Expect(stats).ToNot(BeNil())
	Expect(stats["active_entries"]).ToNot(BeNil())
	Expect(stats["configs"]).ToNot(BeNil())
}

func TestRateLimiterSetConfig(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	config := RateLimitEndpointConfig{
		Requests: 5,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	}

	rl.SetConfig("/custom", config)

	Expect(rl.config["/custom"].Requests).To(Equal(config.Requests))
	Expect(rl.config["/custom"].Window).To(Equal(config.Window))
	Expect(rl.config["/custom"].KeyFunc).ToNot(BeNil())
}

func TestRateLimitMiddleware_NoDoubleCounting(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := telemetry.NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("x-user-id", "123")
		c.Next()
	})
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	var callCountMutex sync.Mutex
	router.POST("/todos", func(c *gin.Context) {
		callCountMutex.Lock()
		callCount++
		callCountMutex.Unlock()
		c.JSON(201, gin.H{"method": "POST", "count": callCount})
	})

	numRequests := 10
	results := make([]int, numRequests)
	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		index := i // Capture loop variable
		wg.Go(func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/todos", strings.NewReader(`{"title":"test"}`))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			remaining, _ := strconv.Atoi(w.Header().Get("X-RateLimit-Remaining"))
			results[index] = remaining
		})
	}

	wg.Wait()

	Expect(callCount).To(Equal(numRequests))

	expectedRemaining := []int{19, 18, 17, 16, 15, 14, 13, 12, 11, 10}
	sort.Ints(results)
	sort.Ints(expectedRemaining)

	Expect(results).To(Equal(expectedRemaining),
		"Concurrent requests should have correct remaining counts without double counting: %v", results)
}
