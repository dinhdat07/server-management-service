package impl

import (
	"context"

	"server-management-service/internal/modules/monitoring/repository"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"gorm.io/gorm"
)

type gormMonitoringRepository struct {
	db *gorm.DB
}

func NewGormMonitoringRepository(db *gorm.DB) repository.MonitoringRepository {
	return &gormMonitoringRepository{
		db: db,
	}
}

func (r *gormMonitoringRepository) UpdateServerStatus(ctx context.Context, serverID string, newStatus serverDomain.ServerStatus, consecutiveFailures int) error {
	return r.db.WithContext(ctx).Model(&serverDomain.Server{}).
		Where("server_id = ?", serverID).
		Updates(map[string]interface{}{
			"current_status":       newStatus,
			"consecutive_failures": consecutiveFailures,
		}).Error
}
