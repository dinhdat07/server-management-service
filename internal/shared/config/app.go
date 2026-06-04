package config

import (
	"fmt"
	"strings"
)

type LoggerConfig struct {
	Level  string
	Format string
}

type Config struct {
	GRPCPort string
	HTTPPort string

	DBUrl         string
	JWTSecret     string
	JWTAccessTTL  int
	RefreshTTL    int
	Port          string
	Env           string
	AdminEmail    string
	AdminPassword string
	ApiBaseUrl    string
	FrontEndUrl         string
	TelegramBotUsername string

	Logger LoggerConfig
}

func Load() (*Config, error) {
	// load .env into os env
	loadEnv()

	accessTTL, err := getEnvInt("JWT_ACCESS_TTL", 3600)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := getEnvInt("JWT_REFRESH_TTL", 604800)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	jwtSecret := getEnvDefault("JWT_SECRET", "")
	if jwtSecret == "" {
		// Backward compatibility with older env key.
		jwtSecret = getEnvDefault("JWT_SECRET_KEY", "")
	}

	httpPort := strings.TrimSpace(getEnvDefault("HTTP_PORT", ""))
	if httpPort == "" {
		httpPort = getEnvDefault("PORT", "8000")
	}

	cfg := &Config{
		GRPCPort:      getEnvDefault("GRPC_PORT", "50051"),
		HTTPPort:      httpPort,
		DBUrl:         getEnvDefault("DB_URL", ""),
		JWTSecret:     jwtSecret,
		JWTAccessTTL:  accessTTL,
		RefreshTTL:    refreshTTL,
		Port:          getEnvDefault("PORT", httpPort),
		Env:           getEnvDefault("ENV", "development"),
		AdminEmail:    getEnvDefault("ADMIN_EMAIL", ""),
		AdminPassword: getEnvDefault("ADMIN_PASSWORD", ""),
		ApiBaseUrl:          getEnvDefault("API_BASE_URL", ""),
		FrontEndUrl:         getEnvDefault("FRONTEND_BASE_URL", ""),
		TelegramBotUsername: getEnvDefault("TELEGRAM_BOT_USERNAME", "YourBotUsername"),
		Logger: LoggerConfig{
			Level:  getEnvDefault("LOG_LEVEL", "info"),
			Format: getEnvDefault("LOG_FORMAT", "text"),
		},
	}

	return cfg, nil
}
