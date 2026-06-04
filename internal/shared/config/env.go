package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

func loadEnv() {
	loadEnvOnce.Do(func() {
		envFile := os.Getenv("ENV_FILE")
		if envFile != "" {
			_ = godotenv.Load(envFile)
			return
		}

		_ = godotenv.Load()
	})
}

func getEnvDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getEnvBool(key string, fallback bool) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	return strconv.ParseBool(val)
}

func getEnvInt(key string, fallback int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	return strconv.Atoi(val)
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	return time.ParseDuration(val)
}

func getEnvFloat(key string, fallback float64) (float64, error) {
	val := os.Getenv(key)
	if val == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(val, 64)
}
