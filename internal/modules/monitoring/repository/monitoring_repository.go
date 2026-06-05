package repository

import (
	"context"

	"server-management-service/internal/modules/monitoring/domain"
	serverDomain "server-management-service/internal/modules/server_management/domain"
)

// MonitoringRepository defines the interface for interacting with the monitoring storage layer.
type MonitoringRepository interface {
	SaveTransitionAndUpdateServer(ctx context.Context, event *domain.StatusTransitionEvent, newStatus serverDomain.ServerStatus) error
}
