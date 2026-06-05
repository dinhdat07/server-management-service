package domain

import "time"

// Current state of a server, stored in Redis.
type MonitoringState struct {
	ServerID           string    `json:"server_id"`
	CurrentRetryCount  int       `json:"current_retry_count"`
	LastCheckTimestamp time.Time `json:"last_check_timestamp"`
}

// Immutable state change event stored in Postgres and synced to Elasticsearch.
type StatusTransitionEvent struct {
	EventID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"event_id"`
	ServerID       string    `gorm:"type:uuid;index;not null" json:"server_id"`
	PreviousStatus string    `gorm:"type:varchar(20);not null" json:"previous_status"`
	CurrentStatus  string    `gorm:"type:varchar(20);not null" json:"current_status"`
	Reason         string    `gorm:"type:varchar(255)" json:"reason"`
	OccurredAt     time.Time `gorm:"autoCreateTime" json:"occurred_at"`
}

func (StatusTransitionEvent) TableName() string {
	return "monitoring_schema.sms_status_transition_logs"
}
