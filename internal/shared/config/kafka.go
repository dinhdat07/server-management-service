package config

import (
	"fmt"
	"os"
	"strings"
)

type KafkaConfig struct {
	Brokers                    []string
	NotificationRequestedTopic string
	ServerTopic                string
	ConsumerGroup              string
}

func LoadKafkaConfig() (KafkaConfig, error) {
	brokers := getEnv("KAFKA_BROKERS", "kafka:9092")

	cfg := KafkaConfig{
		Brokers: splitAndTrim(brokers),

		NotificationRequestedTopic: getEnv("KAFKA_NOTIFICATION_REQUESTED_TOPIC", "notification.requested"),
		ServerTopic:                getEnv("KAFKA_SERVER_TOPIC", "sms.management_schema.servers"),

		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "portal-server-management-group"),
	}

	if len(cfg.Brokers) == 0 {
		return KafkaConfig{}, fmt.Errorf("KAFKA_BROKERS is required")
	}

	return cfg, nil
}
func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
