package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ObservationLog struct {
	ServerID  string    `json:"server_id"`
	IsSuccess bool      `json:"is_success"`
	Timestamp time.Time `json:"timestamp"`
}

type ObservationLogger interface {
	LogObservation(ctx context.Context, serverID string, isSuccess bool) error
}

type observationLogger struct {
	client *elasticsearch.TypedClient
	index  string
}

func NewObservationLogger(client *elasticsearch.TypedClient, index string) ObservationLogger {
	return &observationLogger{
		client: client,
		index:  index,
	}
}

func (l *observationLogger) LogObservation(ctx context.Context, serverID string, isSuccess bool) error {
	log := ObservationLog{
		ServerID:  serverID,
		IsSuccess: isSuccess,
		Timestamp: time.Now().UTC(),
	}

	body, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = l.client.Index(l.index).
		Request(body).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index observation log: %w", err)
	}

	return nil
}
