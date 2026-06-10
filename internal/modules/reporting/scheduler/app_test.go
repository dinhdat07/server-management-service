package scheduler

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/mock"

	"server-management-service/internal/modules/reporting/domain"
)

type mockReportingSvc struct {
	mock.Mock
}

func (m *mockReportingSvc) RequestReport(ctx context.Context, email, startDate, endDate string) error {
	args := m.Called(ctx, email, startDate, endDate)
	return args.Error(0)
}

type mockReportingWorker struct {
	mock.Mock
}

func (m *mockReportingWorker) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *mockReportingWorker) Stop() {
	m.Called()
}

func (m *mockReportingWorker) EnqueueReport(req *domain.ReportRequest) {
	m.Called(req)
}

func TestApp_CronJob(t *testing.T) {
	svc := new(mockReportingSvc)
	worker := new(mockReportingWorker)
	
	app := &App{
		cron:             cron.New(),
		reportingService: svc,
		reportingWorker:  worker,
		adminEmail:       "admin@test.com",
	}

	err := app.setupCronJobs()
	if err != nil {
		t.Fatal(err)
	}

	// Trigger the cron job
	svc.On("RequestReport", mock.Anything, "admin@test.com", mock.Anything, mock.Anything).Return(nil).Once()
	for _, entry := range app.cron.Entries() {
		entry.Job.Run()
	}

	svc.AssertExpectations(t)
}

func TestApp_StartStop(t *testing.T) {
	svc := new(mockReportingSvc)
	worker := new(mockReportingWorker)
	
	app := &App{
		cron:             cron.New(),
		reportingService: svc,
		reportingWorker:  worker,
	}

	worker.On("Start", mock.Anything).Return().Once()
	worker.On("Stop").Return().Once()

	app.Start()
	app.Stop()

	worker.AssertExpectations(t)
}
