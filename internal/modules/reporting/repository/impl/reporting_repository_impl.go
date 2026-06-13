package impl

import (
	"context"

	"server-management-service/internal/modules/reporting/domain"
	"server-management-service/internal/modules/reporting/repository"

	"gorm.io/gorm"
)

type gormReportingRepository struct {
	db *gorm.DB
}

func NewGormReportingRepository(db *gorm.DB) repository.ReportingRepository {
	return &gormReportingRepository{
		db: db,
	}
}

func (r *gormReportingRepository) CreateReportRequest(ctx context.Context, req *domain.ReportRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormReportingRepository) UpdateReportStatus(ctx context.Context, reqID string, status string) error {
	return r.db.WithContext(ctx).Model(&domain.ReportRequest{}).Where("id = ?", reqID).Update("status", status).Error
}

func (r *gormReportingRepository) GetServerCountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Table("management_schema.servers")
	if status != "" {
		query = query.Where("current_status = ?", status)
	}
	err := query.Count(&count).Error
	return count, err
}
