package shared

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestNewResponseCache(t *testing.T) {
	RegisterTestingT(t)

	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())

	cache := NewResponseCache(logger, metrics)

	Expect(cache).ToNot(BeNil())
	Expect(cache.cache).ToNot(BeNil())
	Expect(cache.config).ToNot(BeNil())
	Expect(cache.logger).ToNot(BeNil())
	Expect(cache.metrics).ToNot(BeNil())

	Expect(cache.config).To(HaveKey("/todos"))
	Expect(cache.config).To(HaveKey("default"))

	todosConfig := cache.config["/todos"]
	Expect(todosConfig.TTL).To(Equal(3 * time.Second))
	Expect(todosConfig.Enabled).To(BeTrue())
}

func TestResponseCacheMiddleware_CacheDisabled(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	cache.SetConfig("/test", ResponseCacheConfig{
		TTL:     1 * time.Second,
		Enabled: false,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(BeEmpty())

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
	Expect(w2.Header().Get("X-Cache")).To(BeEmpty())
}

func TestResponseCacheMiddleware_CacheMiss(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w, req)

	Expect(w.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w.Header().Get("X-Cache")).To(Equal("MISS"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	Expect(err).ToNot(HaveOccurred())
	Expect(response).To(HaveKeyWithValue("message", "test"))
	Expect(response).To(HaveKeyWithValue("count", float64(1)))
}

func TestResponseCacheMiddleware_CacheHit(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(Equal("MISS"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w2.Header().Get("X-Cache")).To(Equal("HIT"))

	Expect(w1.Body.String()).To(Equal(w2.Body.String()))
}

func TestResponseCacheMiddleware_CacheExpiration(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	cache.SetConfig("/test", ResponseCacheConfig{
		TTL:     10 * time.Millisecond,
		Enabled: true,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/test", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(Equal("MISS"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w2.Header().Get("X-Cache")).To(Equal("HIT"))

	time.Sleep(20 * time.Millisecond)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
	Expect(w3.Header().Get("X-Cache")).To(Equal("MISS"))
}

func TestResponseCacheMiddleware_DifferentQueryParams(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		cursor := c.Query("cursor")
		c.JSON(200, gin.H{"message": "test", "cursor": cursor, "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(Equal("MISS"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos?cursor=test", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
	Expect(w2.Header().Get("X-Cache")).To(Equal("MISS"))

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/todos?cursor=test", nil)
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
	Expect(w3.Header().Get("X-Cache")).To(Equal("HIT"))
}

func TestResponseCacheMiddleware_NonGETRequests(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.POST("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(201, gin.H{"message": "created", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer([]byte(`{"title":"test"}`)))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(201))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(BeEmpty())

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer([]byte(`{"title":"test2"}`)))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(201))
	Expect(callCount).To(Equal(2))
	Expect(w2.Header().Get("X-Cache")).To(BeEmpty())
}

func TestResponseCacheMiddleware_ErrorResponses(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(500, gin.H{"error": "internal server error", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(500))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(BeEmpty())

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(500))
	Expect(callCount).To(Equal(2))
	Expect(w2.Header().Get("X-Cache")).To(BeEmpty())
}

func TestInvalidateCache(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("x-user-id", 123)
		c.Next()
	})
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(Equal("MISS"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w2.Header().Get("X-Cache")).To(Equal("HIT"))

	cache.InvalidateAllCache()

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w3, req3)

	Expect(w3.Code).To(Equal(200))
	Expect(callCount).To(Equal(2))
	Expect(w3.Header().Get("X-Cache")).To(Equal("MISS"))
}

func TestInvalidateAllCache(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/todos", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "test", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/todos", nil)
	router.ServeHTTP(w1, req1)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos?cursor=test", nil)
	router.ServeHTTP(w2, req2)

	Expect(callCount).To(Equal(2))

	cache.InvalidateAllCache()

	stats := cache.GetStats()
	Expect(stats).To(HaveKeyWithValue("active_entries", 0))
}

func TestResponseCacheGetStats(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	stats := cache.GetStats()
	Expect(stats).To(HaveKey("active_entries"))
	Expect(stats).To(HaveKey("configs"))
	Expect(stats).To(HaveKeyWithValue("active_entries", 0))
	Expect(stats).To(HaveKeyWithValue("configs", 2))
}

func TestResponseCacheSetConfig(t *testing.T) {
	RegisterTestingT(t)
	logger := zap.NewNop()
	metrics := NewAppMetrics(prometheus.NewRegistry())
	cache := NewResponseCache(logger, metrics)

	newConfig := ResponseCacheConfig{
		TTL:     5 * time.Second,
		Enabled: true,
	}
	cache.SetConfig("/custom", newConfig)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cache.CacheMiddleware())

	callCount := 0
	router.GET("/custom", func(c *gin.Context) {
		callCount++
		c.JSON(200, gin.H{"message": "custom", "count": callCount})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/custom", nil)
	router.ServeHTTP(w1, req1)

	Expect(w1.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w1.Header().Get("X-Cache")).To(Equal("MISS"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/custom", nil)
	router.ServeHTTP(w2, req2)

	Expect(w2.Code).To(Equal(200))
	Expect(callCount).To(Equal(1))
	Expect(w2.Header().Get("X-Cache")).To(Equal("HIT"))
}
