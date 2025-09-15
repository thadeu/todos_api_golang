package shared

import "time"

type CursorConfig struct {
	CursorSecretKey string
}

// AppConfig configurações gerais da aplicação
type AppConfig struct {
	// Rate Limiting
	RateLimitEnabled bool
	RateLimitConfigs map[string]RateLimitConfig

	// Response Cache
	CacheEnabled bool
	CacheConfigs map[string]ResponseCacheConfig

	// HTTPS Enforcement
	EnforceHTTPS bool

	// Environment
	Environment string
}

// RateLimitConfig configuração para rate limiting
type RateLimitConfig struct {
	Requests int           // Número de requests permitidos
	Window   time.Duration // Janela de tempo
}

// GetDefaultConfig retorna configuração padrão
func GetDefaultConfig() *AppConfig {
	return &AppConfig{
		RateLimitEnabled: true,
		RateLimitConfigs: map[string]RateLimitConfig{
			"/signup": {
				Requests: 5,
				Window:   time.Minute,
			},
			"/auth": {
				Requests: 10,
				Window:   time.Minute,
			},
			"/todos": {
				Requests: 100,
				Window:   time.Minute,
			},
		},
		CacheEnabled: true,
		CacheConfigs: map[string]ResponseCacheConfig{
			"/todos": {
				TTL:     3 * time.Second,
				Enabled: true,
			},
		},
		EnforceHTTPS: false,
		Environment:  "development",
	}
}
