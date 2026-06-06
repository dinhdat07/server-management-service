package impl

import (
	"context"

	"gorm.io/gorm"

	"server-management-service/internal/modules/reporting/domain"
	"server-management-service/internal/modules/reporting/repository"
)

type GormOutboxRepository struct {
	db *gorm.DB
}

func NewGormOutboxRepository(db *gorm.DB) repository.OutboxRepository {
	return &GormOutboxRepository{
		db: db,
	}
}

func (r *GormOutboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *GormOutboxRepository) CreateReportRequestWithOutbox(ctx context.Context, req *domain.ReportRequest, event *domain.OutboxEvent) error {
	db := r.getDB(ctx)

	// Create Report Request
	if err := db.Create(req).Error; err != nil {
		return err
	}

	// Create Outbox Event
	if err := db.Create(event).Error; err != nil {
		return err
	}

	return nil
}

func (r *GormOutboxRepository) FetchPendingEvents(ctx context.Context, batchSize int) ([]*domain.OutboxEvent, error) {
	db := r.getDB(ctx)
	var events []*domain.OutboxEvent

	err := db.
		Where("status = ?", domain.OutboxStatusPending).
		Order("created_at ASC").
		Limit(batchSize).
		Find(&events).Error

	return events, err
}

func (r *GormOutboxRepository) MarkEventPublished(ctx context.Context, eventID string) error {
	db := r.getDB(ctx)
	
	return db.
		Model(&domain.OutboxEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       domain.OutboxStatusPublished,
			"published_at": db.NowFunc(),
		}).Error
}
