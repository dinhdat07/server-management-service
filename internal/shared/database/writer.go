package database

import (
	"time"

	"github.com/segmentio/kafka-go"
)

func NewWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		MaxAttempts:  5,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
}
