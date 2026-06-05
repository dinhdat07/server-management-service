package impl

import (
	"context"

	"server-management-service/internal/modules/monitoring/domain"
	"server-management-service/internal/modules/monitoring/repository"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"gorm.io/gorm"
)

type GormMonitoringRepository struct {
	db *gorm.DB
}

func NewGormMonitoringRepository(db *gorm.DB) repository.MonitoringRepository {
	return &GormMonitoringRepository{db: db}
}

func (r *GormMonitoringRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *GormMonitoringRepository) SaveTransitionAndUpdateServer(ctx context.Context, event *domain.StatusTransitionEvent, newStatus serverDomain.ServerStatus) error {
	db := r.getDB(ctx)

	serverUpdateResult := db.Model(&serverDomain.Server{}).
		Where("server_id = ?", event.ServerID).
		Updates(map[string]interface{}{
			"current_status":       newStatus,
			"consecutive_failures": 0,
		})
	if serverUpdateResult.Error != nil {
		return serverUpdateResult.Error
	}

	if err := db.Create(event).Error; err != nil {
		return err
	}

	return nil
}
