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
	
	reportingimpl "server-management-service/internal/modules/reporting/repository/impl"
	reportingsvc "server-management-service/internal/modules/reporting/service"
	reportinggrpc "server-management-service/internal/modules/reporting/handler/grpcserver"

	"server-management-service/internal/modules/notification/infrastructure/smtp"
	notificationsvc "server-management-service/internal/modules/notification/service"
	"server-management-service/internal/modules/identity/domain"
	"server-management-service/internal/infrastructure/security"
	authgrpc "server-management-service/internal/modules/identity/handler/grpcserver"
	authrepo "server-management-service/internal/modules/identity/repository/impl"
	authsvc "server-management-service/internal/modules/identity/service"
	
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

	// AutoMigrate Identity
	db.AutoMigrate(
		&domain.User{},
		&domain.AuthSession{},
		&domain.RefreshToken{},
	)

	// Seeder
	seedUsers(db, cfg.AdminEmail, cfg.AdminPassword, cfg.UserEmail, cfg.UserPassword)

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
	reportingWorker := reportingsvc.NewReportingWorker(reportingRepo, esClient, esCfg.ServerIndex, cfg.Reporting.WorkerCount, cfg.Reporting.JobQueueSize, notificationService)
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
		Config:           cfg,
		DB:               db,
		RedisClient:      redisClient,
		ESClient:         esClient,
		ServerHandler:    serverHandler,
		ReportingHandler: reportingHandler,
		ReportingWorker:  reportingWorker,
		AuthHandler:      authHandler,
		NotificationService: notificationService,
	}, nil
}

func seedUsers(db *gorm.DB, adminEmail, adminPassword, userEmail, userPassword string) {
	if adminEmail == "" || adminPassword == "" {
		log.Println("Admin credentials not set, skipping admin seeder.")
	} else {
		seedSingleUser(db, adminEmail, adminPassword, "ADMIN")
	}

	if userEmail == "" || userPassword == "" {
		log.Println("User credentials not set, skipping user seeder.")
	} else {
		seedSingleUser(db, userEmail, userPassword, "USER")
	}
}

func seedSingleUser(db *gorm.DB, email, password, role string) {
	var count int64
	db.Model(&domain.User{}).Where("email = ?", email).Count(&count)
	if count == 0 {
		log.Printf("Seeding default %s user...", role)
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		db.Create(&domain.User{Email: email, Password: string(hash), RoleCode: role})
	}
}
