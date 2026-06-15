package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"server-management-service/internal/modules/reporting/domain"
	domainmock "server-management-service/internal/modules/reporting/domain/mock"
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
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier)

	worker.Start(context.Background())
	time.Sleep(100 * time.Millisecond) // Let worker start
	worker.Stop()
}

func TestNewReportingWorker_InvalidConfigs(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 0, 0, notifier)
	assert.NotNil(t, worker)
}

func TestReportingWorker_ProcessReport_Success(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier).(*reportingWorkerImpl)

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

	uptimeCalc.On("CalculateUptime", mock.Anything, req.StartTime, req.EndTime).Return(80.0, nil).Once()

	notifier.On("SendReportEmail", mock.Anything, "test@test.com", "Server Status Report", mock.Anything).Return(nil).Once()

	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusCompleted).Return(nil).Once()

	worker.processReport(context.Background(), req, 1)

	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
	uptimeCalc.AssertExpectations(t)
}

func TestReportingWorker_ProcessReport_DBError(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier).(*reportingWorkerImpl)

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
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier).(*reportingWorkerImpl)

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

func TestReportingWorker_doWork_OnlineOfflineErr(t *testing.T) {
	repo := new(mockReportRepo)
	worker := NewReportingWorker(repo, nil, 1, 10, nil).(*reportingWorkerImpl)
	req := &domain.ReportRequest{}

	repo.On("GetServerCountByStatus", mock.Anything, "").Return(int64(10), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "ONLINE").Return(int64(0), errors.New("online err")).Once()
	err := worker.doWork(context.Background(), req)
	assert.Error(t, err)

	repo.On("GetServerCountByStatus", mock.Anything, "").Return(int64(10), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "ONLINE").Return(int64(8), nil).Once()
	repo.On("GetServerCountByStatus", mock.Anything, "OFFLINE").Return(int64(0), errors.New("offline err")).Once()
	err = worker.doWork(context.Background(), req)
	assert.Error(t, err)
}

func TestReportingWorker_doWork_UptimeErr(t *testing.T) {
	repo := new(mockReportRepo)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, nil).(*reportingWorkerImpl)
	req := &domain.ReportRequest{}

	repo.On("GetServerCountByStatus", mock.Anything, mock.Anything).Return(int64(10), nil).Times(3)
	uptimeCalc.On("CalculateUptime", mock.Anything, req.StartTime, req.EndTime).Return(0.0, errors.New("uptime err")).Once()
	
	err := worker.doWork(context.Background(), req)
	assert.Error(t, err)
}

func TestReportingWorker_doWork_EmailErr(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier).(*reportingWorkerImpl)
	req := &domain.ReportRequest{}

	repo.On("GetServerCountByStatus", mock.Anything, mock.Anything).Return(int64(10), nil).Times(3)
	uptimeCalc.On("CalculateUptime", mock.Anything, req.StartTime, req.EndTime).Return(100.0, nil).Once()
	notifier.On("SendReportEmail", mock.Anything, req.RequestorEmail, "Server Status Report", mock.Anything).Return(errors.New("smtp err")).Once()

	err := worker.doWork(context.Background(), req)
	assert.Error(t, err)
}

func TestReportingWorker_EnqueueAndProcess(t *testing.T) {
	repo := new(mockReportRepo)
	notifier := new(mockNotifier)
	uptimeCalc := new(domainmock.MockUptimeCalculator)
	worker := NewReportingWorker(repo, uptimeCalc, 1, 10, notifier)

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
	uptimeCalc.On("CalculateUptime", mock.Anything, req.StartTime, req.EndTime).Return(80.0, nil).Once()
	notifier.On("SendReportEmail", mock.Anything, "test@test.com", "Server Status Report", mock.Anything).Return(nil).Once()
	repo.On("UpdateReportStatus", mock.Anything, req.ID.String(), domain.ReportStatusCompleted).Return(nil).Once()

	worker.Start(context.Background())
	worker.EnqueueReport(req)
	time.Sleep(200 * time.Millisecond) // Give time to process
	worker.Stop()
}
