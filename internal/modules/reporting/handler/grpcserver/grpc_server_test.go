package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	reportingv1 "server-management-service/gen/go/reporting/v1"
)

type mockReportingSvc struct {
	mock.Mock
}

func (m *mockReportingSvc) RequestReport(ctx context.Context, email, startDate, endDate string) error {
	args := m.Called(ctx, email, startDate, endDate)
	return args.Error(0)
}

func TestReportingGrpcHandler_RequestReport(t *testing.T) {
	svc := new(mockReportingSvc)
	h := NewReportingGrpcHandler(svc)

	t.Run("empty email", func(t *testing.T) {
		req := &reportingv1.RequestReportRequest{TargetEmail: ""}
		_, err := h.RequestReport(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		svc.On("RequestReport", mock.Anything, "admin@test.com", "2026-01-01", "2026-01-31").
			Return(nil).Once()

		req := &reportingv1.RequestReportRequest{
			TargetEmail: "admin@test.com",
			StartDate:   "2026-01-01",
			EndDate:     "2026-01-31",
		}
		resp, err := h.RequestReport(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int32(202), resp.Code)
		assert.Equal(t, "processing", resp.Status)
	})

	t.Run("service error", func(t *testing.T) {
		svc.On("RequestReport", mock.Anything, "admin@test.com", "", "").
			Return(errors.New("db error")).Once()

		req := &reportingv1.RequestReportRequest{
			TargetEmail: "admin@test.com",
		}
		_, err := h.RequestReport(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}
