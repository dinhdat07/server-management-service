package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	ReportStatusPending    = "PENDING"
	ReportStatusProcessing = "PROCESSING"
	ReportStatusCompleted  = "COMPLETED"
	ReportStatusFailed     = "FAILED"

	OutboxStatusPending   = "PENDING"
	OutboxStatusPublished = "PUBLISHED"
	OutboxStatusFailed    = "FAILED"
)

// ReportRequest represents an asynchronous request to generate a server uptime report.
type ReportRequest struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	RequestorEmail string    `gorm:"type:varchar(255);not null"`
	StartTime      time.Time `gorm:"not null"`
	EndTime        time.Time `gorm:"not null"`
	Status         string    `gorm:"type:varchar(50);not null"`
	CorrelationID  string    `gorm:"type:varchar(255);not null;index"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (ReportRequest) TableName() string {
	return "reporting_schema.report_requests"
}

// OutboxEvent represents an event to be published asynchronously using the Transactional Outbox pattern.
type OutboxEvent struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	AggregateType string         `gorm:"type:varchar(100);not null"`
	AggregateID   string         `gorm:"type:varchar(255);not null"`
	EventType     string         `gorm:"type:varchar(100);not null"`
	Payload       datatypes.JSON `gorm:"type:jsonb;not null"`
	Status        string         `gorm:"type:varchar(50);not null;default:'PENDING';index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	PublishedAt   *time.Time
}

func (OutboxEvent) TableName() string {
	return "integration_schema.outbox_events"
}

// ReportPayload represents the exact JSON structure stored inside OutboxEvent.Payload
type ReportPayload struct {
	RequestID      string `json:"request_id"`
	RequestorEmail string `json:"requestor_email"`
	StartTimeUnix  int64  `json:"start_time_unix"`
	EndTimeUnix    int64  `json:"end_time_unix"`
	CorrelationID  string `json:"correlation_id"`
}

// NewReportRequest is a factory function to create a new report request and its corresponding outbox event
func NewReportRequest(requestorEmail string, startTime, endTime time.Time, correlationID string) (*ReportRequest, *OutboxEvent, error) {
	requestID := uuid.New()
	
	req := &ReportRequest{
		ID:             requestID,
		RequestorEmail: requestorEmail,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         ReportStatusPending,
		CorrelationID:  correlationID,
	}

	payload := ReportPayload{
		RequestID:      requestID.String(),
		RequestorEmail: requestorEmail,
		StartTimeUnix:  startTime.Unix(),
		EndTimeUnix:    endTime.Unix(),
		CorrelationID:  correlationID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	event := &OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "ReportRequest",
		AggregateID:   requestID.String(),
		EventType:     "server.report.requested",
		Payload:       payloadBytes,
		Status:        OutboxStatusPending,
	}

	return req, event, nil
}
