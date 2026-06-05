package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"server-management-service/internal/modules/monitoring/domain"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockMonitoringRepo struct {
	mock.Mock
}

func (m *mockMonitoringRepo) SaveTransitionAndUpdateServer(ctx context.Context, event *domain.StatusTransitionEvent, newStatus serverDomain.ServerStatus) error {
	args := m.Called(ctx, event, newStatus)
	return args.Error(0)
}

type mockTxManager struct{}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestEvaluate_FailureThreshold(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-123"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	txManager := &mockTxManager{}
	service := NewMonitoringService(repo, db, txManager)

	// Test 1: First Failure (Online -> Online)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "0",
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 1).SetVal(1)

	// Repo should NOT be called since status didn't change
	err := service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertNotCalled(t, "SaveTransitionAndUpdateServer")

	// Test 2: Second Failure (Online -> Offline)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "ONLINE",
		"retry_count": "1", // 1 previous failure
	})
	mockRedis.ExpectHSet(redisKey, "status", "OFFLINE", "retry_count", 0).SetVal(1) // final state
	
	repo.On("SaveTransitionAndUpdateServer", ctx, mock.AnythingOfType("*domain.StatusTransitionEvent"), serverDomain.ServerStatusOffline).Return(nil)

	err = service.Evaluate(ctx, serverID, "1.1.1.1", false)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertCalled(t, "SaveTransitionAndUpdateServer", ctx, mock.Anything, serverDomain.ServerStatusOffline)
}

func TestEvaluate_RecoveryThreshold(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-456"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	txManager := &mockTxManager{}
	service := NewMonitoringService(repo, db, txManager)

	// Test: Offline -> Online (Recovery)
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "OFFLINE",
		"retry_count": "5", // Multiple failures in offline state
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1) // final state

	repo.On("SaveTransitionAndUpdateServer", ctx, mock.AnythingOfType("*domain.StatusTransitionEvent"), serverDomain.ServerStatusOnline).Return(nil)

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	repo.AssertCalled(t, "SaveTransitionAndUpdateServer", ctx, mock.Anything, serverDomain.ServerStatusOnline)
}

func TestEvaluate_DBError(t *testing.T) {
	ctx := context.Background()
	serverID := "svr-789"
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	db, mockRedis := redismock.NewClientMock()
	repo := new(mockMonitoringRepo)
	txManager := &mockTxManager{}
	service := NewMonitoringService(repo, db, txManager)

	// Test: Offline -> Online (Recovery) but DB fails
	mockRedis.ExpectHGetAll(redisKey).SetVal(map[string]string{
		"status":      "OFFLINE",
		"retry_count": "0",
	})
	mockRedis.ExpectHSet(redisKey, "status", "ONLINE", "retry_count", 0).SetVal(1)

	expectedErr := errors.New("db error")
	repo.On("SaveTransitionAndUpdateServer", ctx, mock.Anything, serverDomain.ServerStatusOnline).Return(expectedErr)

	err := service.Evaluate(ctx, serverID, "1.1.1.1", true)
	assert.ErrorIs(t, err, expectedErr)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}
