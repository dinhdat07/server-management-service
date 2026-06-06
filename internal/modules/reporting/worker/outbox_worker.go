package worker

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"server-management-service/internal/modules/reporting/repository"
)

type OutboxWorker struct {
	repo         repository.OutboxRepository
	kafkaWriter  *kafka.Writer
	tickInterval time.Duration
	batchSize    int
}

func NewOutboxWorker(repo repository.OutboxRepository, kafkaWriter *kafka.Writer) *OutboxWorker {
	return &OutboxWorker{
		repo:         repo,
		kafkaWriter:  kafkaWriter,
		tickInterval: 5 * time.Second,
		batchSize:    100,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) error {
	log.Println("Starting Outbox Worker...")
	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox Worker stopped")
			return nil
		case <-ticker.C:
			w.processOutbox(ctx)
		}
	}
}

func (w *OutboxWorker) processOutbox(ctx context.Context) {
	events, err := w.repo.FetchPendingEvents(ctx, w.batchSize)
	if err != nil {
		log.Printf("Failed to fetch outbox events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("Processing %d outbox events", len(events))

	for _, event := range events {
		msg := kafka.Message{
			Key:   []byte(event.AggregateID),
			Value: event.Payload,
		}

		// Write to Kafka
		if err := w.kafkaWriter.WriteMessages(ctx, msg); err != nil {
			log.Printf("Failed to publish event %s to Kafka: %v", event.ID, err)
			continue
		}

		// Mark as published
		if err := w.repo.MarkEventPublished(ctx, event.ID.String()); err != nil {
			log.Printf("Failed to mark event %s as published: %v", event.ID, err)
		}
	}
}
