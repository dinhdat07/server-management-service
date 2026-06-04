package database

import (
	"time"

	"github.com/segmentio/kafka-go"
)

func NewReader(brokers []string, topic []string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupTopics: topic,
		GroupID:     groupID,

		MinBytes: 1,
		MaxBytes: 10e6,

		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
		MaxWait:        1 * time.Second,
	})
}
