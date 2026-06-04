package main

import (
	"context"
	"log"

	"server-management-service/internal/infrastructure/cdc"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
)

func main() {
	kafkaCfg, err := config.LoadKafkaConfig()
	if err != nil {
		log.Fatalf("failed to load kafka config: %v", err)
	}

	esCfg := config.LoadElasticsearchConfig()

	esClient, err := database.NewElasticsearchClient(context.Background(), []string{esCfg.URL})
	if err != nil {
		log.Fatalf("elasticsearch connection failed: %v", err)
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

	consumer, err := cdc.New(context.Background(), kafkaCfg, esCfg, esClient, redisClient)
	if err != nil {
		log.Fatalf("init cdc consumer: %v", err)
	}

	if err := consumer.Run(); err != nil {
		log.Fatalf("run cdc consumer: %v", err)
	}
}
