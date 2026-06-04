package database

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func PingKafka(ctx context.Context, brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	dialer := &kafka.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", brokers[0], err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close kafka ping connection broker=%s error=%v", brokers[0], err)
		}
	}()

	log.Printf("kafka broker connection verified broker=%s", brokers[0])
	return nil
}
