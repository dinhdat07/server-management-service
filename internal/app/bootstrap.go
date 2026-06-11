package app

import (
	"context"
	"log"
	"server-management-service/internal/modules/server_management/handler/grpcserver"
	"server-management-service/internal/modules/server_management/repository/impl"
	"server-management-service/internal/modules/server_management/service"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"

	infraRedis "server-management-service/internal/infrastructure/redis"
	"server-management-service/internal/infrastructure/storage"

	reportinggrpc "server-management-service/internal/modules/reporting/handler/grpcserver"
	reportingimpl "server-management-service/internal/modules/reporting/repository/impl"
	reportingsvc "server-management-service/internal/modules/reporting/service"

	"server-management-service/internal/infrastructure/ratelimit"

	"buf.build/go/protovalidate"

	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/infrastructure/security"
	authgrpc "server-management-service/internal/modules/identity/handler/grpcserver"
	authrepo "server-management-service/internal/modules/identity/repository/impl"
	authsvc "server-management-service/internal/modules/identity/service"
	"server-management-service/internal/modules/notification/infrastructure/smtp"
	notificationsvc "server-management-service/internal/modules/notification/service"
)

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := database.GetInstance(cfg.DBUrl)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return nil, err
	}

	storage.AutoMigrate(db)
	storage.SeedUsers(db, cfg.AdminEmail, cfg.AdminPassword, cfg.UserEmail, cfg.UserPassword)

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		log.Printf("failed to load redis config: %v", err)
	}

	redisClient := database.NewRedisClient(redisCfg)
	if redisClient != nil {
		if err := database.PingRedis(context.Background(), redisClient); err != nil {
			log.Printf("redis ping failed: %v", err)
		}
	}

	esCfg := config.LoadElasticsearchConfig()
	esClient, err := database.NewElasticsearchClient(context.Background(), []string{esCfg.URL})
	if err != nil {
		log.Printf("elasticsearch connection failed: %v", err)
	}

	rateLimitCfg, err := config.LoadRateLimitConfig()
	if err != nil {
		log.Printf("failed to load rate limit config: %v", err)
	}

	var rateLimiter ratelimit.Limiter
	var rateLimitKeyBuilder ratelimit.KeyBuilder
	if rateLimitCfg != nil && rateLimitCfg.Enabled {
		rateLimiter, err = ratelimit.NewRedisLimiter(redisClient)
		if err != nil {
			log.Printf("failed to initialize rate limiter: %v", err)
		} else {
			rateLimitKeyBuilder = ratelimit.NewKeyBuilder(rateLimitCfg.Prefix)
		}
	}

	validator, err := protovalidate.New()
	if err != nil {
		log.Printf("failed to initialize protovalidate: %v", err)
	}
	csrfManager := security.NewCSRFManager()

	serverRepo := impl.NewGormServerRepository(db)
	serverCache := infraRedis.NewServerCache(redisClient)
	serverSvc := service.NewServerService(serverRepo, serverCache)
	serverHandler := grpcserver.NewServerManagementServer(serverSvc)

	smtpConfig := smtp.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		UseAuth:  cfg.SMTP.UseAuth,
		UseTLS:   cfg.SMTP.UseTLS,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		FromName: cfg.SMTP.FromName,
	}
	smtpMailer := smtp.NewMailer(smtpConfig)
	notificationService := notificationsvc.NewNotificationService(smtpMailer)

	reportingRepo := reportingimpl.NewGormReportingRepository(db)
	uptimeCalc := elasticsearch.NewESUptimeCalculator(esClient, esCfg.ServerIndex)
	reportingWorker := reportingsvc.NewReportingWorker(reportingRepo, uptimeCalc, cfg.Reporting.WorkerCount, cfg.Reporting.JobQueueSize, notificationService)
	reportingService := reportingsvc.NewReportingService(reportingRepo, reportingWorker)
	reportingHandler := reportinggrpc.NewReportingGrpcHandler(reportingService)

	// Identity Service
	userRepo := authrepo.NewUserRepository(db)
	sessionRepo := authrepo.NewAuthSessionRepository(db)
	refreshRepo := authrepo.NewRefreshTokenRepository(db)
	revoStore := authrepo.NewSessionRevocationStore(redisClient)
	tokenMgr := security.NewTokenManager(cfg.JWTSecret)

	authService := authsvc.NewAuthService(userRepo, sessionRepo, refreshRepo, revoStore, tokenMgr)
	authHandler := authgrpc.NewAuthServer(authService)

	return &App{
		Config:              cfg,
		DB:                  db,
		RedisClient:         redisClient,
		ESClient:            esClient,
		ServerHandler:       serverHandler,
		ReportingHandler:    reportingHandler,
		ReportingWorker:     reportingWorker,
		AuthHandler:         authHandler,
		NotificationService: notificationService,

		Validator:           validator,
		Authenticator:       security.NewAuthenticator(cfg.JWTSecret, redisClient),
		Authorizer:          security.NewAuthorizer(),
		CSRFManager:         csrfManager,
		RateLimiter:         rateLimiter,
		RateLimitKeyBuilder: rateLimitKeyBuilder,
		RateLimitConfig:     rateLimitCfg,
	}, nil
}
