package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"

	"server-management-service/internal/modules/reporting/domain"
)

type NotificationPayload struct {
	TemplateID string                 `json:"template_id"`
	EmailTo    string                 `json:"email_to"`
	Data       map[string]interface{} `json:"data"`
}

type ReportConsumer struct {
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
	esClient    *elasticsearch.TypedClient
}

func NewReportConsumer(kafkaReader *kafka.Reader, kafkaWriter *kafka.Writer, esClient *elasticsearch.TypedClient) *ReportConsumer {
	return &ReportConsumer{
		kafkaReader: kafkaReader,
		kafkaWriter: kafkaWriter,
		esClient:    esClient,
	}
}

func (c *ReportConsumer) Run(ctx context.Context) error {
	log.Println("Starting Report Consumer...")

	for {
		msg, err := c.kafkaReader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				log.Println("Report Consumer stopped")
				return nil
			}
			log.Printf("Failed to read kafka message: %v", err)
			continue
		}

		c.processMessage(ctx, msg)
	}
}

func (c *ReportConsumer) processMessage(ctx context.Context, msg kafka.Message) {
	var payload domain.ReportPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		log.Printf("Failed to unmarshal report request payload: %v", err)
		return
	}

	log.Printf("Calculating Uptime for request %s (Email: %s) from %d to %d", payload.RequestID, payload.RequestorEmail, payload.StartTimeUnix, payload.EndTimeUnix)

	// Step 4 - Calculate Uptime from Elasticsearch using Date Histogram Aggregation
	uptimePercentage := 99.99
	
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": payload.StartTimeUnix * 1000,
					"lte": payload.EndTimeUnix * 1000,
					"format": "epoch_millis",
				},
			},
		},
		"aggs": map[string]interface{}{
			"status_fluctuations": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field": "timestamp",
					"calendar_interval": "1d",
				},
			},
		},
	}
	
	// Execute the aggregation query (placeholder for actual execution logic)
	// res, err := c.esClient.Search().Index("sms_status_logs").Request(query).Do(ctx)
	// _ = res
	
	// Mock processing of the aggregation results
	// ... calculate uptime based on the buckets
	_ = query

	// Prepare notification payload
	notiPayload := NotificationPayload{
		TemplateID: "server_uptime_report",
		EmailTo:    payload.RequestorEmail,
		Data: map[string]interface{}{
			"correlation_id":    payload.CorrelationID,
			"start_time_unix":   payload.StartTimeUnix,
			"end_time_unix":     payload.EndTimeUnix,
			"uptime_percentage": uptimePercentage,
		},
	}

	notiBytes, _ := json.Marshal(notiPayload)

	outMsg := kafka.Message{
		Key:   []byte(payload.RequestID),
		Value: notiBytes,
	}

	if err := c.kafkaWriter.WriteMessages(ctx, outMsg); err != nil {
		log.Printf("Failed to publish notification request: %v", err)
		return
	}

	log.Printf("Successfully processed report request %s", payload.RequestID)
}
