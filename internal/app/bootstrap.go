package app

import (
	"context"
	"log"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
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

	return &App{
		Config:      cfg,
		DB:          db,
		RedisClient: redisClient,
		ESClient:    esClient,
	}, nil
}
