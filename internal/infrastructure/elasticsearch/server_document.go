package elasticsearch

import (
	"time"
)

type ServerDocument struct {
	ID                  string    `json:"id"` // Maps to server_id
	ServerName          string    `json:"server_name"`
	IPv4                string    `json:"ipv4"`
	CurrentStatus       string    `json:"current_status"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
