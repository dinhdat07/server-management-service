package config

import (
	"fmt"
	"strings"
	"time"
)

type RateLimitConfig struct {
	Enabled bool
	Prefix  string

	DefaultLimit  int
	DefaultBurst  int
	DefaultPeriod time.Duration

	FailOpen bool
}

func LoadRateLimitConfig() (*RateLimitConfig, error) {
	enabled, err := getEnvBool("RATE_LIMIT_ENABLED", false)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_ENABLED: %w", err)
	}

	defaultLimit, err := getEnvInt("RATE_LIMIT_DEFAULT_LIMIT", 300)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_DEFAULT_LIMIT: %w", err)
	}

	defaultBurst, err := getEnvInt("RATE_LIMIT_DEFAULT_BURST", defaultLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_DEFAULT_BURST: %w", err)
	}

	defaultPeriod, err := getEnvDuration("RATE_LIMIT_DEFAULT_PERIOD", time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_DEFAULT_PERIOD: %w", err)
	}

	failOpen, err := getEnvBool("RATE_LIMIT_FAIL_OPEN", true)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_FAIL_OPEN: %w", err)
	}

	cfg := &RateLimitConfig{
		Enabled:       enabled,
		Prefix:        strings.TrimSpace(getEnvDefault("RATE_LIMIT_PREFIX", "portal:rl")),
		DefaultLimit:  defaultLimit,
		DefaultBurst:  defaultBurst,
		DefaultPeriod: defaultPeriod,
		FailOpen:      failOpen,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *RateLimitConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("rate limit config is nil")
	}

	if !c.Enabled {
		return nil
	}

	if c.Prefix == "" {
		return fmt.Errorf("RATE_LIMIT_PREFIX is required when rate limit is enabled")
	}

	if c.DefaultLimit <= 0 {
		return fmt.Errorf("RATE_LIMIT_DEFAULT_LIMIT must be > 0")
	}

	if c.DefaultBurst <= 0 {
		return fmt.Errorf("RATE_LIMIT_DEFAULT_BURST must be > 0")
	}

	if c.DefaultPeriod <= 0 {
		return fmt.Errorf("RATE_LIMIT_DEFAULT_PERIOD must be > 0")
	}

	return nil
}
