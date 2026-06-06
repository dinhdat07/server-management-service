package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"server-management-service/internal/modules/reporting/domain"
	"server-management-service/internal/modules/reporting/repository"
)

type ReportingService interface {
	RequestReport(ctx context.Context, email string, startDate string, endDate string) error
}

type reportingServiceImpl struct {
	repo      repository.OutboxRepository
	txManager repository.TxManager
}

func NewReportingService(repo repository.OutboxRepository, txManager repository.TxManager) ReportingService {
	return &reportingServiceImpl{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *reportingServiceImpl) RequestReport(ctx context.Context, email string, startDate string, endDate string) error {
	// Parse dates
	// Format is YYYY-MM-DD
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return err
	}
	
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return err
	}
	// Make sure end is at the end of the day
	end = end.Add(24 * time.Hour).Add(-time.Nanosecond)

	correlationID := uuid.New().String()

	req, event, err := domain.NewReportRequest(email, start, end, correlationID)
	if err != nil {
		return err
	}

	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		return s.repo.CreateReportRequestWithOutbox(txCtx, req, event)
	})
}
