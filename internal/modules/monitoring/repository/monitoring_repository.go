package repository

import (
	"context"

	serverDomain "server-management-service/internal/modules/server_management/domain"
)

// MonitoringRepository defines the interface for interacting with the monitoring storage layer.
type MonitoringRepository interface {
	UpdateServerStatus(ctx context.Context, serverID string, newStatus serverDomain.ServerStatus, consecutiveFailures int) error
}
