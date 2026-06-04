package cdc

import (
	"context"
	"fmt"
	"log"

	"server-management-service/internal/shared/config"
	esinfra "server-management-service/internal/infrastructure/elasticsearch"
	redisinfra "server-management-service/internal/infrastructure/redis"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func New(ctx context.Context, kafkaCfg config.KafkaConfig, esCfg config.ElasticsearchConfig, esClient *esv8.Client, redisClient redis.UniversalClient) (*Consumer, error) {
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
