package repository

import (
	"context"

	"server-management-service/internal/modules/reporting/domain"
)

type OutboxRepository interface {
	CreateReportRequestWithOutbox(ctx context.Context, req *domain.ReportRequest, event *domain.OutboxEvent) error
	FetchPendingEvents(ctx context.Context, batchSize int) ([]*domain.OutboxEvent, error)
	MarkEventPublished(ctx context.Context, eventID string) error
}
