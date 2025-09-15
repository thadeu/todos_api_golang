package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestNewRateLimiter(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())

	rl := NewRateLimiter(logger, metrics)

	Expect(rl).ToNot(BeNil())
	Expect(rl.cache).ToNot(BeNil())
	Expect(rl.config).ToNot(BeNil())
	Expect(rl.logger).ToNot(BeNil())
	Expect(rl.metrics).ToNot(BeNil())

	Expect(rl.config).To(HaveKey("/signup"))
	Expect(rl.config).To(HaveKey("/auth"))
	Expect(rl.config).To(HaveKey("/todos"))
	Expect(rl.config).To(HaveKey("default"))

	signupConfig := rl.config["/signup"]
	Expect(signupConfig.Requests).To(Equal(100))
	Expect(signupConfig.Window).To(Equal(time.Second))
}

func TestRateLimitMiddleware_AllowedRequests(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 3,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(callCount).To(Equal(i))

		Expect(w.Header().Get("X-RateLimit-Limit")).To(Equal("3"))
		expectedRemaining := strconv.Itoa(3 - i)
		Expect(w.Header().Get("X-RateLimit-Remaining")).To(Equal(expectedRemaining))
		Expect(w.Header().Get("X-RateLimit-Reset")).ToNot(BeEmpty())
	}
}

func TestRateLimitMiddleware_ExceedLimit(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 2,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	for i := 1; i <= 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(callCount).To(Equal(i))
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w, req)

	Expect(w.Code).To(Equal(429))
	Expect(callCount).To(Equal(2))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	Expect(err).ToNot(HaveOccurred())
	Expect(response).To(HaveKeyWithValue("error", "Rate limit exceeded"))
	Expect(response).To(HaveKey("message"))
	Expect(response["message"]).To(ContainSubstring("Too many requests"))
	Expect(response).To(HaveKey("retry_after"))

	Expect(w.Header().Get("X-RateLimit-Limit")).To(Equal("2"))
	Expect(w.Header().Get("X-RateLimit-Remaining")).To(Equal("0"))
	Expect(w.Header().Get("X-RateLimit-Reset")).ToNot(BeEmpty())
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(429))
	Expect(callCount).To(Equal(2))
}

func TestRateLimitMiddleware_UserBasedLimiting(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  getUserID,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("x-user-id", userID)
		}
		c.Next()
	})
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-User-ID", "123")
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-User-ID", "456")
	req2.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-User-ID", "123")
	req3.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(429))
	Expect(callCount).To(Equal(2))
}

func TestRateLimitMiddleware_WindowReset(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 1,
		Window:   50 * time.Millisecond,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(429))
	Expect(callCount).To(Equal(1))

	time.Sleep(60 * time.Millisecond)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
}

func TestRateLimitMiddleware_DefaultConfig(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/unknown", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/unknown", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w, req)

	Expect(w.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	defaultConfig := rl.config["default"]
	Expect(w.Header().Get("X-RateLimit-Limit")).To(Equal(strconv.Itoa(defaultConfig.Requests)))
	Expect(w.Header().Get("X-RateLimit-Remaining")).ToNot(BeEmpty())
	Expect(w.Header().Get("X-RateLimit-Reset")).ToNot(BeEmpty())
}

func TestRateLimitMiddleware_MultipleEndpoints(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/endpoint1", RateLimitEndpointConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})
	rl.SetConfig("/endpoint2", RateLimitEndpointConfig{
		Requests: 2,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount1, callCount2 := 0, 0
	router.GET("/endpoint1", func(c *gin.Context) {
		callCount1++
		c.JSON(200, gin.H{"endpoint": 1, "count": callCount1})
	})
	router.GET("/endpoint2", func(c *gin.Context) {
		callCount2++
		c.JSON(200, gin.H{"endpoint": 2, "count": callCount2})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/endpoint1", nil)
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount1).To(Equal(1))
	Expect(w1.Header().Get("X-RateLimit-Limit")).To(Equal("1"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/endpoint2", nil)
	req2.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount2).To(Equal(1))
	Expect(w2.Header().Get("X-RateLimit-Limit")).To(Equal("2"))

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/endpoint1", nil)
	req3.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(429))
	Expect(callCount1).To(Equal(1))

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/endpoint2", nil)
	req4.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w4, req4)

	Expect(w4.Code).To(Equal(200))
	Expect(callCount2).To(Equal(2))
}

func TestGetUserID(t *testing.T) {
	RegisterTestingT(t)
	gin.SetMode(gin.TestMode)

	c1, _ := gin.CreateTestContext(httptest.NewRecorder())
	c1.Set("x-user-id", 123)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "127.0.0.1:8080"

	userID1 := getUserID(c1)
	Expect(userID1).To(Equal("user_123"))

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.1:8080"

	userID2 := getUserID(c2)
	Expect(userID2).To(Equal("192.168.1.1"))
}

func TestRateLimiterSetConfig(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	newConfig := RateLimitEndpointConfig{
		Requests: 5,
		Window:   30 * time.Second,
		KeyFunc:  GetClientIP,
	}
	rl.SetConfig("/custom", newConfig)

	Expect(rl.config).To(HaveKey("/custom"))
	config := rl.config["/custom"]
	Expect(config.Requests).To(Equal(5))
	Expect(config.Window).To(Equal(30 * time.Second))
}

func TestRateLimiterGetStats(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	stats := rl.GetStats()
	Expect(stats).To(HaveKey("active_entries"))
	Expect(stats).To(HaveKey("configs"))
	Expect(stats).To(HaveKeyWithValue("active_entries", 0))
	Expect(stats).To(HaveKeyWithValue("configs", 4))
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	req2.RemoteAddr = "192.168.1.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(429))
	Expect(callCount).To(Equal(1))
}

func TestRateLimitMiddleware_RealIPHeader(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	rl := NewRateLimiter(logger, metrics)

	rl.SetConfig("/test", RateLimitEndpointConfig{
		Requests: 1,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "success", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Real-IP", "203.0.113.1")
	req1.RemoteAddr = "127.0.0.1:8080"
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Real-IP", "203.0.113.1")
	req2.RemoteAddr = "192.168.1.1:8080"
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(429))
	Expect(callCount).To(Equal(1))
}
