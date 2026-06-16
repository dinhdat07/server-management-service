package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	serverDomain "server-management-service/internal/modules/server_management/domain"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockMonitoringRepo struct {
	mock.Mock
}

func (m *mockMonitoringRepo) UpdateServerStatus(ctx context.Context, serverID string, newStatus serverDomain.ServerStatus, consecutiveFailures int) error {
	args := m.Called(ctx, serverID, newStatus, consecutiveFailures)
	return args.Error(0)
}

type mockObservationLogger struct {
	mock.Mock
}

func (m *mockObservationLogger) LogObservation(ctx context.Context, serverID string, isSuccess bool) error {
	args := m.Called(ctx, serverID, isSuccess)
	return args.Error(0)
}

func TestEvaluate_FailureThreshold(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-123"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	
	// Set threshold to 2 for testing
	service := NewMonitoringService(repo, db, esLogger, 2)

	// Test 1: First Failure (Online -> Online)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "0",
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 1).SetVal(1)

	// ES Logger should be called on every ping
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	
	// Repo should be called to update consecutive_failures even though it's still online
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 1).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)

	// Test 2: Second Failure (Online -> Offline)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "1", // 1 previous failure
	})
	mockRedis.ExpectHSet(redisKey, "status", "OFFLINE", "retry_count", 0).SetVal(1) // final state
	
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOffline, 0).Return(nil).Once()

	err = service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
}

func TestEvaluate_RecoveryThreshold(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-456"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	// Test: Offline -> Online (Recovery)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "OFFLINE",
		"retry_count": "5", // Multiple failures in offline state
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1) // final state

	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
}

func TestEvaluate_DBError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-789"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	// Test: Offline -> Online (Recovery) but DB fails
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "OFFLINE",
		"retry_count": "0",
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1)

	expectedErr := errors.New("db error")
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 0).Return(expectedErr).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.ErrorIs(t, err, expectedErr)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
}
