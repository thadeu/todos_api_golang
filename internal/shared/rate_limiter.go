package shared

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// RateLimitEndpointConfig configuração específica para rate limiting por endpoint
type RateLimitEndpointConfig struct {
	Requests int                       // Número de requests permitidos
	Window   time.Duration             // Janela de tempo
	KeyFunc  func(*gin.Context) string // Função para gerar chave única
}

// RateLimiter estrutura para gerenciar rate limiting
type RateLimiter struct {
	cache   *cache.Cache
	config  map[string]RateLimitEndpointConfig
	logger  *zap.Logger
	metrics *AppMetrics
}

// RateLimitEntry entrada no cache para rate limiting
type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

// NewRateLimiter cria uma nova instância do rate limiter
func NewRateLimiter(logger *zap.Logger, metrics *AppMetrics) *RateLimiter {
	// Cache com limpeza automática a cada 5 minutos
	c := cache.New(5*time.Minute, 10*time.Minute)

	// Configurações padrão por endpoint
	configs := map[string]RateLimitEndpointConfig{
		"/signup": {
			Requests: 5, // 5 tentativas de signup por minuto
			Window:   time.Minute,
			KeyFunc:  getClientIP,
		},
		"/auth": {
			Requests: 10, // 10 tentativas de login por minuto
			Window:   time.Minute,
			KeyFunc:  getClientIP,
		},
		"/todos": {
			Requests: 100, // 100 requests por minuto para todos
			Window:   time.Minute,
			KeyFunc:  getUserID,
		},
		"default": {
			Requests: 60, // 60 requests por minuto para outros endpoints
			Window:   time.Minute,
			KeyFunc:  getClientIP,
		},
	}

	return &RateLimiter{
		cache:   c,
		config:  configs,
		logger:  logger,
		metrics: metrics,
	}
}

// RateLimitMiddleware middleware para rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Buscar configuração para o endpoint
		config, exists := rl.config[path]
		if !exists {
			// Usar configuração padrão
			config = rl.config["default"]
		}

		// Gerar chave única para o rate limiting
		key := rl.generateKey(c, path, config.KeyFunc)

		// Verificar rate limit
		allowed, remaining, resetTime, err := rl.checkRateLimit(key, config)
		if err != nil {
			rl.logger.Error("Rate limit check failed",
				zap.String("key", key),
				zap.String("path", path),
				zap.Error(err))
			c.Next()
			return
		}

		// Determinar tipo de chave para métricas
		keyType := "ip"
		// Verificar se a chave contém "user_" para determinar o tipo
		if strings.Contains(key, "user_") {
			keyType = "user"
		}

		// Adicionar headers informativos
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// Registrar métrica de rate limit hit
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

		// Registrar métrica de request permitido
		if rl.metrics != nil {
			rl.metrics.RecordRateLimitAllowed(c.Request.Context(), path, keyType)
		}

		c.Next()
	}
}

// checkRateLimit verifica se o request está dentro do limite
func (rl *RateLimiter) checkRateLimit(key string, config RateLimitEndpointConfig) (bool, int, time.Time, error) {
	now := time.Now()

	// Buscar entrada no cache
	if entry, found := rl.cache.Get(key); found {
		rateLimitEntry := entry.(RateLimitEntry)

		// Se ainda está na janela de tempo
		if now.Before(rateLimitEntry.ResetTime) {
			if rateLimitEntry.Count >= config.Requests {
				return false, 0, rateLimitEntry.ResetTime, nil
			}

			// Incrementar contador
			rateLimitEntry.Count++
			rl.cache.Set(key, rateLimitEntry, cache.DefaultExpiration)

			return true, config.Requests - rateLimitEntry.Count, rateLimitEntry.ResetTime, nil
		}
	}

	// Criar nova entrada ou resetar contador
	resetTime := now.Add(config.Window)
	newEntry := RateLimitEntry{
		Count:     1,
		ResetTime: resetTime,
	}

	rl.cache.Set(key, newEntry, config.Window)

	return true, config.Requests - 1, resetTime, nil
}

// generateKey gera chave única para rate limiting
func (rl *RateLimiter) generateKey(c *gin.Context, path string, keyFunc func(*gin.Context) string) string {
	identifier := keyFunc(c)
	return fmt.Sprintf("rate_limit:%s:%s", path, identifier)
}

// getClientIP extrai IP do cliente
func getClientIP(c *gin.Context) string {
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

// getUserID extrai ID do usuário autenticado
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("x-user-id"); exists {
		return fmt.Sprintf("user_%v", userID)
	}

	// Fallback para IP se não estiver autenticado
	return getClientIP(c)
}

// SetConfig permite configurar rate limits para endpoints específicos
func (rl *RateLimiter) SetConfig(path string, config RateLimitEndpointConfig) {
	rl.config[path] = config
}

// GetStats retorna estatísticas do rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Contar entradas ativas no cache
	activeEntries := rl.cache.ItemCount()

	stats["active_entries"] = activeEntries
	stats["configs"] = len(rl.config)

	return stats
}
