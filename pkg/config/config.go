package config

import (
	"time"
	response "todoapp/pkg/response"
)

type CursorConfig struct {
	CursorSecretKey string
}

type AppConfig struct {
	RateLimitEnabled bool
	RateLimitConfigs map[string]RateLimitConfig

	CacheEnabled bool
	CacheConfigs map[string]response.ResponseCacheConfig

	EnforceHTTPS bool

	Environment string
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

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
		CacheConfigs: map[string]response.ResponseCacheConfig{
			"/todos": {
				TTL:     3 * time.Second,
				Enabled: true,
			},
		},
		EnforceHTTPS: false,
		Environment:  "development",
	}
}
