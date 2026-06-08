package config

import (
	"fmt"
	"strings"
)

type LoggerConfig struct {
	Level  string
	Format string
}

type SMTPConfig struct {
	Host     string
	Port     string
	UseAuth  bool
	UseTLS   bool
	Username string
	Password string
	From     string
	FromName string
}

type ReportingConfig struct {
	WorkerCount  int
	JobQueueSize int
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
	UserEmail     string
	UserPassword  string
	ApiBaseUrl    string
	FrontEndUrl         string
	TelegramBotUsername string

	Logger LoggerConfig
	SMTP   SMTPConfig
	Reporting ReportingConfig
}

func Load() (*Config, error) {
	// load .env into os env
	loadEnv()

	accessTTL, err := GetEnvInt("JWT_ACCESS_TTL", 3600)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := GetEnvInt("JWT_REFRESH_TTL", 604800)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	jwtSecret := GetEnvDefault("JWT_SECRET", "")
	if jwtSecret == "" {
		// Backward compatibility with older env key.
		jwtSecret = GetEnvDefault("JWT_SECRET_KEY", "")
	}

	httpPort := strings.TrimSpace(GetEnvDefault("HTTP_PORT", ""))
	if httpPort == "" {
		httpPort = GetEnvDefault("PORT", "8000")
	}

	workerCount, err := GetEnvInt("REPORTING_WORKER_COUNT", 5)
	if err != nil {
		workerCount = 5
	}
	
	jobQueueSize, err := GetEnvInt("REPORTING_JOB_QUEUE_SIZE", 100)
	if err != nil {
		jobQueueSize = 100
	}

	cfg := &Config{
		GRPCPort:      GetEnvDefault("GRPC_PORT", "50051"),
		HTTPPort:      httpPort,
		DBUrl:         GetEnvDefault("DB_URL", ""),
		JWTSecret:     jwtSecret,
		JWTAccessTTL:  accessTTL,
		RefreshTTL:    refreshTTL,
		Port:          GetEnvDefault("PORT", httpPort),
		Env:           GetEnvDefault("ENV", "development"),
		AdminEmail:    GetEnvDefault("ADMIN_EMAIL", ""),
		AdminPassword: GetEnvDefault("ADMIN_PASSWORD", ""),
		UserEmail:     GetEnvDefault("USER_EMAIL", ""),
		UserPassword:  GetEnvDefault("USER_PASSWORD", ""),
		ApiBaseUrl:          GetEnvDefault("API_BASE_URL", ""),
		FrontEndUrl:         GetEnvDefault("FRONTEND_BASE_URL", ""),
		TelegramBotUsername: GetEnvDefault("TELEGRAM_BOT_USERNAME", "YourBotUsername"),
		Logger: LoggerConfig{
			Level:  GetEnvDefault("LOG_LEVEL", "info"),
			Format: GetEnvDefault("LOG_FORMAT", "text"),
		},
		SMTP: SMTPConfig{
			Host:     GetEnvDefault("SMTP_HOST", "localhost"),
			Port:     GetEnvDefault("SMTP_PORT", "1025"),
			UseAuth:  GetEnvDefault("SMTP_USE_AUTH", "false") == "true",
			UseTLS:   GetEnvDefault("SMTP_USE_TLS", "false") == "true",
			Username: GetEnvDefault("SMTP_USERNAME", ""),
			Password: GetEnvDefault("SMTP_PASSWORD", ""),
			From:     GetEnvDefault("SMTP_FROM", "no-reply@vcs.com"),
			FromName: GetEnvDefault("SMTP_FROM_NAME", "VCS Server Management"),
		},
		Reporting: ReportingConfig{
			WorkerCount:  workerCount,
			JobQueueSize: jobQueueSize,
		},
	}

	return cfg, nil
}
