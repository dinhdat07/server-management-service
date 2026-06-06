package app

import (
	"context"
	"log"
	"server-management-service/internal/modules/server_management/handler/grpcserver"
	"server-management-service/internal/modules/server_management/repository/impl"
	"server-management-service/internal/modules/server_management/service"
	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
	
	reportingimpl "server-management-service/internal/modules/reporting/repository/impl"
	reportingsvc "server-management-service/internal/modules/reporting/service"
	reportinggrpc "server-management-service/internal/modules/reporting/handler/grpcserver"
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
	serverSearcher := elasticsearch.NewServerSearcher(esClient, esCfg.ServerIndex)
	serverSvc := service.NewServerService(serverRepo, serverSearcher)
	serverHandler := grpcserver.NewServerManagementServer(serverSvc)

	reportingTxManager := reportingimpl.NewGormTxManager(db)
	reportingRepo := reportingimpl.NewGormOutboxRepository(db)
	reportingService := reportingsvc.NewReportingService(reportingRepo, reportingTxManager)
	reportingHandler := reportinggrpc.NewReportingGrpcHandler(reportingService)

	return &App{
		Config:           cfg,
		DB:               db,
		RedisClient:      redisClient,
		ESClient:         esClient,
		ServerHandler:    serverHandler,
		ReportingHandler: reportingHandler,
	}, nil
}
