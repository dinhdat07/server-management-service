package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewReportRequest(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	req, err := NewReportRequest("admin@example.com", start, end, "corr-123")

	assert.NoError(t, err)
	assert.Equal(t, "admin@example.com", req.RequestorEmail)
	assert.Equal(t, "corr-123", req.CorrelationID)
	assert.Equal(t, ReportStatusPending, req.Status)
	assert.NotEqual(t, "", req.ID.String())
}

func TestReportRequest_TableName(t *testing.T) {
	r := &ReportRequest{}
	assert.Equal(t, "reporting_schema.report_requests", r.TableName())
}

func TestReportStatus_Constants(t *testing.T) {
	assert.Equal(t, "PENDING", ReportStatusPending)
	assert.Equal(t, "PROCESSING", ReportStatusProcessing)
	assert.Equal(t, "COMPLETED", ReportStatusCompleted)
	assert.Equal(t, "FAILED", ReportStatusFailed)
}
