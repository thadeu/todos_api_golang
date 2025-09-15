package shared

import "time"

type CursorConfig struct {
	CursorSecretKey string
}

// AppConfig general application configurations
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

// RateLimitConfig configuration for rate limiting
type RateLimitConfig struct {
	Requests int
	Window   time.Duration // Janela de tempo
}

// GetDefaultConfig returns default configuration
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
