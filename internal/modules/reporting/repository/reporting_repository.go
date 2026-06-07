package repository

import (
	"context"

	"server-management-service/internal/modules/reporting/domain"
)

type ReportingRepository interface {
	CreateReportRequest(ctx context.Context, req *domain.ReportRequest) error
	UpdateReportStatus(ctx context.Context, id string, status string) error
	GetServerCountByStatus(ctx context.Context, status string) (int64, error)
}
