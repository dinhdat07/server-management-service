package service

import (
	"context"
	"errors"
	"testing"

	"server-management-service/internal/modules/reporting/domain"
	repomock "server-management-service/internal/modules/reporting/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorker struct {
	mock.Mock
}

func (m *mockWorker) Start(ctx context.Context) {}
func (m *mockWorker) Stop()                     {}
func (m *mockWorker) EnqueueReport(req *domain.ReportRequest) {
	m.Called(req)
}

func TestRequestReport_Success(t *testing.T) {
	ctx := context.Background()
	repo := repomock.NewMockReportingRepository(t)
	worker := new(mockWorker)

	svc := NewReportingService(repo, worker)

	repo.On("CreateReportRequest", ctx, mock.AnythingOfType("*domain.ReportRequest")).Return(nil).Once()
	worker.On("EnqueueReport", mock.AnythingOfType("*domain.ReportRequest")).Return().Once()

	err := svc.RequestReport(ctx, "admin@example.com", "2026-06-01", "2026-06-02")
	assert.NoError(t, err)
}

func TestRequestReport_InvalidStartDate(t *testing.T) {
	ctx := context.Background()
	repo := repomock.NewMockReportingRepository(t)
	worker := new(mockWorker)

	svc := NewReportingService(repo, worker)
	err := svc.RequestReport(ctx, "admin@example.com", "invalid-date", "2026-06-02")
	assert.Error(t, err)
}

func TestRequestReport_InvalidEndDate(t *testing.T) {
	ctx := context.Background()
	repo := repomock.NewMockReportingRepository(t)
	worker := new(mockWorker)

	svc := NewReportingService(repo, worker)
	err := svc.RequestReport(ctx, "admin@example.com", "2026-06-01", "not-a-date")
	assert.Error(t, err)
}

func TestRequestReport_EmptyEmail(t *testing.T) {
	ctx := context.Background()
	repo := repomock.NewMockReportingRepository(t)
	worker := new(mockWorker)

	svc := NewReportingService(repo, worker)

	// Empty email is now rejected by domain validation
	err := svc.RequestReport(ctx, "", "2026-06-01", "2026-06-02")
	assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	worker.AssertNotCalled(t, "EnqueueReport")
}

func TestRequestReport_DBError(t *testing.T) {
	ctx := context.Background()
	repo := repomock.NewMockReportingRepository(t)
	worker := new(mockWorker)

	dbErr := errors.New("db error")
	svc := NewReportingService(repo, worker)

	repo.On("CreateReportRequest", ctx, mock.Anything).Return(dbErr).Once()

	err := svc.RequestReport(ctx, "admin@example.com", "2026-06-01", "2026-06-02")
	assert.ErrorIs(t, err, dbErr)
	worker.AssertNotCalled(t, "EnqueueReport")
}

