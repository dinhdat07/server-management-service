package service

import (
	"context"
	"errors"
	"testing"

	monitoringDomain "server-management-service/internal/modules/monitoring/domain"
	"server-management-service/internal/modules/monitoring/repository"
	mockRepo "server-management-service/internal/modules/monitoring/repository/mock"
	serverDomain "server-management-service/internal/modules/server_management/domain"

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

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	// First Failure (Online -> Online) with threshold=2
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 1).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 1).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
	stateStore.AssertExpectations(t)
}

func TestEvaluate_SecondFailureGoesOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-123"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	// Second consecutive failure crosses threshold (Online -> Offline)
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 1}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "OFFLINE", 0).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOffline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
	stateStore.AssertExpectations(t)
}

func TestEvaluate_RecoveryThreshold(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-456"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	// Test: Offline -> Online (Recovery)
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "OFFLINE", RetryCount: 5}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(nil).Once()

	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
	stateStore.AssertExpectations(t)
}

func TestEvaluate_DBError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-789"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	// Test: Offline -> Online (Recovery) but DB fails
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "OFFLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(nil).Once()

	expectedErr := errors.New("db error")
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOnline, 0).Return(expectedErr).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.ErrorIs(t, err, expectedErr)
	repo.AssertExpectations(t)
	esLogger.AssertExpectations(t)
	stateStore.AssertExpectations(t)
}

func TestEvaluate_OnlineStaysOnline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-keep"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
}

func TestEvaluate_OfflineStaysOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-off"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "OFFLINE", RetryCount: 3}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "OFFLINE", 4).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
}

func TestEvaluate_DefaultStatus(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-default"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
}

func TestEvaluate_GetServerStateError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-redis-err"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	stateStore.On("GetServerState", ctx, serverID).Return(nil, errors.New("redis timeout")).Once()
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis timeout")
}

func TestEvaluate_ESLogErrorNonFatal(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-es-err"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	esLogger.On("LogObservation", ctx, serverID, true).Return(errors.New("ES unavailable")).Once()
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
}

func TestEvaluate_RedisEmpty(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-empty"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	stateStore.On("GetServerState", ctx, serverID).Return((*monitoringDomain.ServerState)(nil), repository.ErrServerStateNotFound).Once()
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.ErrorIs(t, err, repository.ErrServerStateNotFound)
}

func TestEvaluate_RedisHSetError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-hset-err"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 2)

	hsetErr := errors.New("redis write error")
	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "ONLINE", 0).Return(hsetErr).Once()
	esLogger.On("LogObservation", ctx, serverID, true).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.ErrorIs(t, err, hsetErr)
}

func TestEvaluate_ThresholdOne_ImmediateOffline(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-t1"

	stateStore := mockRepo.NewMockServerStateStore(t)
	repo := new(mockMonitoringRepo)
	esLogger := new(mockObservationLogger)
	service := NewMonitoringService(repo, stateStore, esLogger, 1)

	stateStore.On("GetServerState", ctx, serverID).Return(&monitoringDomain.ServerState{Status: "ONLINE", RetryCount: 0}, nil).Once()
	stateStore.On("SetServerState", ctx, serverID, "OFFLINE", 0).Return(nil).Once()
	esLogger.On("LogObservation", ctx, serverID, false).Return(nil).Once()
	repo.On("UpdateServerStatus", ctx, serverID, serverDomain.ServerStatusOffline, 0).Return(nil).Once()

	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
}
