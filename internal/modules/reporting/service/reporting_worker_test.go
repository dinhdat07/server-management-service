package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"server-management-service/internal/modules/reporting/domain"
)

type mockReportRepo struct {
	mock.Mock
}

func (m *mockReportRepo) CreateReportRequest(ctx context.Context, req *domain.ReportRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockReportRepo) UpdateReportStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockReportRepo) GetServerCountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

type mockNotifier struct {
	mock.Mock
}

func (m *mockNotifier) SendReportEmail(ctx context.Context, toEmail string, subject string, htmlBody string) error {
	args := m.Called(ctx, toEmail, subject, htmlBody)
	return args.Error(0)
}

func TestReportingWorker_StartStop(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	worker := NewReportingWorker(repo, nil, "idx", 1, 10, notifier)

	worker.Start(context.Background())
	time.Sleep(100 * time.Millisecond) // Let worker start
	worker.Stop()
}

func TestReportingWorker_ProcessReport_Success(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	worker := NewReportingWorker(repo, nil, "idx", 1, 10, notifier).(*reportingWorkerImpl)

	req := &domain.ReportRequest{
		ID:             uuid.New(),
		RequestorEmail: "test@test.com",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusProcessing).Return(nil).Once()
	
	repo.On("GetServerCountByStatus", mock.Anything, "").Return(int64(10), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "ONLINE").Return(int64(8), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "OFFLINE").Return(int64(2), nil).Once()

	notifier.On("SendReportEmail", mock.Anything, "test@test.com", "Server Status Report", mock.Anything).Return(nil).Once()

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusCompleted).Return(nil).Once()

	worker.processReport(context.Background(), req, 1)

	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestReportingWorker_ProcessReport_DBError(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	worker := NewReportingWorker(repo, nil, "idx", 1, 10, notifier).(*reportingWorkerImpl)

	req := &domain.ReportRequest{
		ID:             uuid.New(),
		RequestorEmail: "test@test.com",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusProcessing).Return(errors.New("db down")).Once()

	// Should abort early
	worker.processReport(context.Background(), req, 1)

	repo.AssertExpectations(t)
}

func TestReportingWorker_ProcessReport_WorkError(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	worker := NewReportingWorker(repo, nil, "idx", 1, 10, notifier).(*reportingWorkerImpl)

	req := &domain.ReportRequest{
		ID:             uuid.New(),
		RequestorEmail: "test@test.com",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusProcessing).Return(nil).Once()
	
	repo.On("GetServerCountByStatus", mock.Anything, "").Return(int64(0), errors.New("query failed")).Once()

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusFailed).Return(nil).Once()

	worker.processReport(context.Background(), req, 1)

	repo.AssertExpectations(t)
}

func TestReportingWorker_EnqueueAndProcess(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	worker := NewReportingWorker(repo, nil, "idx", 1, 10, notifier)

	req := &domain.ReportRequest{
		ID:             uuid.New(),
		RequestorEmail: "test@test.com",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusProcessing).Return(nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "").Return(int64(10), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "ONLINE").Return(int64(8), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "OFFLINE").Return(int64(2), nil).Once()
	notifier.On("SendReportEmail", mock.Anything, "test@test.com", "Server Status Report", mock.Anything).Return(nil).Once()
	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusCompleted).Return(nil).Once()

	worker.Start(context.Background())
	worker.EnqueueReport(req)
	time.Sleep(200 * time.Millisecond) // Give time to process
	worker.Stop()
}
