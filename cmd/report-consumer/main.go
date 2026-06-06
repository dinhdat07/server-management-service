package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"

	"server-management-service/internal/modules/reporting/worker"
	"server-management-service/internal/shared/config"
)

func main() {
	_, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	kafkaBrokers := config.GetEnvDefault("KAFKA_BROKERS", "localhost:9092")
	reportTopic := config.GetEnvDefault("KAFKA_REPORT_REQUESTED_TOPIC", "server.report.requested")
	notificationTopic := config.GetEnvDefault("KAFKA_NOTIFICATION_TOPIC", "portal.notification.requested")
	esURL := config.GetEnvDefault("ELASTICSEARCH_URL", "http://localhost:9200")

	// Initialize Elasticsearch Client
	esCfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}
	esClient, err := elasticsearch.NewTypedClient(esCfg)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Initialize Kafka Reader
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBrokers},
		GroupID:        "report-consumer-group",
		Topic:          reportTopic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	defer kafkaReader.Close()

	// Initialize Kafka Writer
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Topic:    notificationTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Initialize Report Consumer
	reportConsumer := worker.NewReportConsumer(kafkaReader, kafkaWriter, esClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := reportConsumer.Run(ctx); err != nil {
			log.Fatalf("Report consumer failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Report Consumer...")
	cancel()
}
