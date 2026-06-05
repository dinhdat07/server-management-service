package cdc

import (
	"context"
	"fmt"
	"log"

	"server-management-service/internal/shared/config"
	esinfra "server-management-service/internal/infrastructure/elasticsearch"
	redisinfra "server-management-service/internal/infrastructure/redis"
	"server-management-service/internal/shared/database"

	"github.com/segmentio/kafka-go"
)

func New() (*Consumer, error) {
	ctx := context.Background()

	kafkaCfg, err := config.LoadKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kafka config: %w", err)
	}

	esCfg := config.LoadElasticsearchConfig()

	esClient, err := database.NewElasticsearchClient(ctx, []string{esCfg.URL})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connection failed: %w", err)
	}

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		log.Printf("failed to load redis config: %v", err)
	}

	redisClient := database.NewRedisClient(redisCfg)
	if redisClient != nil {
		if err := database.PingRedis(ctx, redisClient); err != nil {
			log.Printf("redis ping failed: %v", err)
		}
	}

	serverIndexer := esinfra.NewServerIndexer(esClient, esCfg.ServerIndex)
	if err := serverIndexer.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure server index: %w", err)
	}

	serverCache := redisinfra.NewServerCache(redisClient)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kafkaCfg.Brokers,
		GroupID:        kafkaCfg.ConsumerGroup,
		Topic:          kafkaCfg.ServerTopic,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0, // synchronous commits
	})

	router := NewRouterHandler(map[string]TopicHandler{
		kafkaCfg.ServerTopic: NewServerEventHandler(serverIndexer, serverCache),
	})

	log.Printf(
		"[CDC Consumer] Configured: brokers=%v topic=%s group=%s elasticsearch=%s index=%s",
		kafkaCfg.Brokers,
		kafkaCfg.ServerTopic,
		kafkaCfg.ConsumerGroup,
		esCfg.URL,
		esCfg.ServerIndex,
	)

	return NewConsumer(reader, router), nil
}
