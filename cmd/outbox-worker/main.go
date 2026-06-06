package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"server-management-service/internal/modules/reporting/repository/impl"
	"server-management-service/internal/modules/reporting/worker"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Database
	db, err := database.GetInstance(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	outboxRepo := impl.NewGormOutboxRepository(db)

	// Initialize Kafka Writer
	kafkaBrokers := config.GetEnvDefault("KAFKA_BROKERS", "localhost:9092")
	outboxTopic := config.GetEnvDefault("KAFKA_OUTBOX_TOPIC", "portal.outbox.events")

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Topic:    outboxTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Initialize Outbox Worker
	outboxWorker := worker.NewOutboxWorker(outboxRepo, kafkaWriter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := outboxWorker.Run(ctx); err != nil {
			log.Fatalf("Outbox worker failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Outbox Worker...")
	cancel()
}
