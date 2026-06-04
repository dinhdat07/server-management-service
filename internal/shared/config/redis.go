package config

import (
	"fmt"
	"time"
)

type RedisConfig struct {
	Enabled      bool
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
}

func LoadRedisConfig() (*RedisConfig, error) {
	enabled, err := getEnvBool("REDIS_ENABLED", false)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_ENABLED: %w", err)
	}

	db, err := getEnvInt("REDIS_DB", 0)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	dialTimeout, err := getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DIAL_TIMEOUT: %w", err)
	}

	readTimeout, err := getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_WRITE_TIMEOUT: %w", err)
	}

	poolSize, err := getEnvInt("REDIS_POOL_SIZE", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_POOL_SIZE: %w", err)
	}

	minIdleConns, err := getEnvInt("REDIS_MIN_IDLE_CONNS", 2)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS: %w", err)
	}

	maxRetries, err := getEnvInt("REDIS_MAX_RETRIES", 3)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MAX_RETRIES: %w", err)
	}

	return &RedisConfig{
		Enabled:      enabled,
		Addr:         getEnvDefault("REDIS_ADDR", ""),
		Password:     getEnvDefault("REDIS_PASSWORD", ""),
		DB:           db,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   maxRetries,
	}, nil
}
