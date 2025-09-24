package config

import (
	"time"
)

type AppConfig struct {
	RateLimitEnabled bool
	RateLimitConfigs map[string]RateLimitConfig

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
		EnforceHTTPS: false,
		Environment:  "development",
	}
}
