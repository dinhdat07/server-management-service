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

func TestEvaluate_FirstFailureStaysOnline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-123"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	// First Failure (Online -> Online) with threshold=2
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "0",
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 1).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 1).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
}

func TestEvaluate_SecondFailureGoesOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-123"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	// Second consecutive failure crosses threshold (Online -> Offline)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "1",
	})
	mockRedis.ExpectHSet(redisKey, "status", "OFFLINE", "retry_count", 0).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOffline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
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

func TestEvaluate_OnlineStaysOnline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-keep"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "ONLINE", "retry_count": "0"})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestEvaluate_OfflineStaysOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-off"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "OFFLINE", "retry_count": "3"})
	mockRedis.ExpectHSet(redisKey, "status", "OFFLINE", "retry_count", 4).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestEvaluate_DefaultStatus(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-default"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "", "retry_count": "0"})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
}

func TestEvaluate_RedisHGetAllError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-redis-err"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	mockRedis.ExpectHGetAll(redisKey).SetErr(errors.New("redis timeout"))
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get server info from redis")
}

func TestEvaluate_ESLogErrorNonFatal(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-es-err"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	esLogger.On("LogObservation", ctx, serverID, true).Return(errors.New("ES unavailable")).Once()
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "ONLINE", "retry_count": "0"})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1)

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
}

func TestEvaluate_RedisEmpty(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-empty"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{})
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server info not found in redis")
}

func TestEvaluate_RedisHSetError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-hset-err"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 2)

	hsetErr := errors.New("redis write error")
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "ONLINE", "retry_count": "0"})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetErr(hsetErr)
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update redis status")
}

func TestEvaluate_ThresholdOne_ImmediateOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-t1"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, db, esLogger, 1)

	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{"status": "ONLINE", "retry_count": "0"})
	mockRedis.ExpectHSet(redisKey, "status", "OFFLINE", "retry_count", 0).SetVal(1)
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOffline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}
