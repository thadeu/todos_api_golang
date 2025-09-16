package response

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	. "todoapp/pkg"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ResponseCacheConfig configuration for response cache
type ResponseCacheConfig struct {
	TTL     time.Duration
	Enabled bool
}

// ResponseCache middleware para cache de respostas
type ResponseCache struct {
	cache   *cache.Cache
	config  map[string]ResponseCacheConfig
	logger  *zap.Logger
	metrics *AppMetrics
}

// CachedResponse estrutura para armazenar resposta em cache
type CachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	Timestamp  time.Time           `json:"timestamp"`
}

// NewResponseCache creates a new response cache instance
func NewResponseCache(logger *zap.Logger, metrics *AppMetrics) *ResponseCache {
	c := cache.New(5*time.Minute, 10*time.Minute)

	configs := map[string]ResponseCacheConfig{
		"/todos": {
			TTL:     3 * time.Second,
			Enabled: true,
		},
		"default": {
			TTL:     1 * time.Second,
			Enabled: true,
		},
	}

	return &ResponseCache{
		cache:   c,
		config:  configs,
		logger:  logger,
		metrics: metrics,
	}
}

// CacheMiddleware middleware para cache de respostas
func (rc *ResponseCache) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		config, exists := rc.config[path]
		if !exists {
			config = rc.config["default"]
		}

		if !config.Enabled {
			c.Next()
			return
		}

		cacheKey := rc.generateCacheKey(c, path)

		// Tentar buscar do cache
		if cachedResp, found := rc.cache.Get(cacheKey); found {
			cached := cachedResp.(CachedResponse)

			if time.Since(cached.Timestamp) < config.TTL {
				// Cache hit - criar span para indicar que veio do cache
				_, span := CreateChildSpan(c.Request.Context(), "cache.response.hit", []attribute.KeyValue{
					attribute.String("cache.key", cacheKey),
					attribute.String("cache.path", path),
					attribute.String("cache.age", time.Since(cached.Timestamp).String()),
					attribute.String("cache.source", "memory"),
				})
				defer span.End()

				// Adicionar atributos de sucesso do cache
				span.SetAttributes(
					attribute.Int("cache.status_code", cached.StatusCode),
					attribute.Int("cache.body_size", len(cached.Body)),
					attribute.String("cache.ttl", config.TTL.String()),
				)

				if rc.metrics != nil {
					rc.metrics.RecordCacheHit(c.Request.Context(), path)
				}

				rc.logger.Debug("Cache hit",
					zap.String("path", path),
					zap.String("cache_key", cacheKey),
					zap.Duration("age", time.Since(cached.Timestamp)))

				// Restaurar headers
				for key, values := range cached.Headers {
					for _, value := range values {
						c.Header(key, value)
					}
				}

				// Adicionar header indicando que veio do cache
				c.Header("X-Cache", "HIT")
				c.Header("X-Cache-Age", fmt.Sprintf("%.0f", time.Since(cached.Timestamp).Seconds()))

				c.Data(cached.StatusCode, "application/json", cached.Body)
				c.Abort()
				return
			} else {
				// Cache expirado - remover
				rc.cache.Delete(cacheKey)
			}
		}

		ctx, span := CreateChildSpan(c.Request.Context(), "cache.response.miss", []attribute.KeyValue{
			attribute.String("cache.key", cacheKey),
			attribute.String("cache.path", path),
			attribute.String("cache.source", "memory"),
		})
		defer span.End()

		if rc.metrics != nil {
			rc.metrics.RecordCacheMiss(c.Request.Context(), path)
		}

		rc.logger.Debug("Cache miss",
			zap.String("path", path),
			zap.String("cache_key", cacheKey))

		// Interceptar a resposta
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Processar request
		c.Next()

		if writer.statusCode >= 200 && writer.statusCode < 300 {
			_, cacheSpan := CreateChildSpan(ctx, "cache.response.store", []attribute.KeyValue{
				attribute.String("cache.key", cacheKey),
				attribute.String("cache.path", path),
				attribute.String("cache.source", "memory"),
				attribute.Int("cache.status_code", writer.statusCode),
				attribute.Int("cache.body_size", len(writer.body.Bytes())),
				attribute.String("cache.ttl", config.TTL.String()),
			})
			cacheSpan.End()

			// Armazenar no cache
			cachedResp := CachedResponse{
				StatusCode: writer.statusCode,
				Headers:    writer.Header(),
				Body:       writer.body.Bytes(),
				Timestamp:  time.Now(),
			}

			rc.cache.Set(cacheKey, cachedResp, config.TTL)

			// Adicionar header indicando que foi cacheado
			c.Header("X-Cache", "MISS")
		}
	}
}

// generateCacheKey generates unique cache key
func (rc *ResponseCache) generateCacheKey(c *gin.Context, path string) string {
	// Incluir path, query parameters e user ID se autenticado
	keyParts := []string{path}

	// Adicionar query parameters
	if c.Request.URL.RawQuery != "" {
		keyParts = append(keyParts, c.Request.URL.RawQuery)
	}

	// Adicionar user ID se autenticado
	if userID, exists := c.Get("x-user-id"); exists {
		keyParts = append(keyParts, fmt.Sprintf("user_%v", userID))
	} else {
		keyParts = append(keyParts, fmt.Sprintf("ip_%s", GetClientIP(c)))
	}

	// Criar hash da chave
	keyString := strings.Join(keyParts, "|")
	hash := md5.Sum([]byte(keyString))

	return fmt.Sprintf("cache:%s:%x", path, hash)
}

// InvalidateCache invalidates cache for specific user
func (rc *ResponseCache) InvalidateCache(userID int, path string) {
	keys := rc.cache.Items()

	for key := range keys {
		if strings.Contains(key, fmt.Sprintf("user_%d", userID)) && strings.Contains(key, path) {
			rc.cache.Delete(key)
			rc.logger.Debug("Cache invalidated",
				zap.String("key", key),
				zap.Int("user_id", userID),
				zap.String("path", path))
		}
	}
}

// InvalidateAllCache invalida todo o cache
func (rc *ResponseCache) InvalidateAllCache() {
	rc.cache.Flush()
	rc.logger.Info("All cache invalidated")
}

// SetConfig allows configuring cache for specific endpoints
func (rc *ResponseCache) SetConfig(path string, config ResponseCacheConfig) {
	rc.config[path] = config
}

// GetStats returns cache statistics
func (rc *ResponseCache) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Contar entradas ativas no cache
	activeEntries := rc.cache.ItemCount()

	stats["active_entries"] = activeEntries
	stats["configs"] = len(rc.config)

	return stats
}

// responseWriter wrapper para interceptar a resposta
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
