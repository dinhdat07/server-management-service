package database

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafkago.Writer
}

func NewPublisher(writer *kafkago.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	if topic == "" {
		return fmt.Errorf("kafka topic is required")
	}

	if key == "" {
		return fmt.Errorf("kafka message key is required")
	}

	if err := p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("publish kafka message: %w", err)
	}

	return nil
}
