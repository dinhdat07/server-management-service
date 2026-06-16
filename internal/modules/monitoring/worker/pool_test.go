package worker

import (
	"context"
	"testing"
	"time"

	infraRedis "server-management-service/internal/infrastructure/redis"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Pinger
type mockPinger struct {
	mock.Mock
}

func (m *mockPinger) Ping(ip string, timeout time.Duration) bool {
	args := m.Called(ip, timeout)
	return args.Bool(0)
}

// Mock MonitoringService
type mockMonitoringService struct {
	mock.Mock
}

func (m *mockMonitoringService) Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error {
	args := m.Called(ctx, serverID, ip, pingSuccess)
	return args.Error(0)
}

func TestWorkerPool_Run(t *testing.T) {
	ctx := context.Background()
	db, mockRedis := redismock.NewClientMock()
	mockRedis.MatchExpectationsInOrder(false)
	monService := new(mockMonitoringService)
	pinger := new(mockPinger)

	pool := NewWorkerPool(db, monService, pinger, 2, 1*time.Second)

	serverIDs := []string{"id-1", "id-2"}

	// 1. Mock Fetch all Server IDs
	mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetVal(serverIDs)

	// 2. Mock processing each server
	mockRedis.ExpectHGet("server:info:id-1", "ipv4").SetVal("1.1.1.1")
	pinger.On("Ping", "1.1.1.1", 1*time.Second).Return(true)
	monService.On("Evaluate", mock.Anything, "id-1", "1.1.1.1", true).Return(nil)

	mockRedis.ExpectHGet("server:info:id-2", "ipv4").SetVal("2.2.2.2")
	pinger.On("Ping", "2.2.2.2", 1*time.Second).Return(false)
	monService.On("Evaluate", mock.Anything, "id-2", "2.2.2.2", false).Return(nil)

	err := pool.Run(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
	pinger.AssertExpectations(t)
	monService.AssertExpectations(t)
}

func TestWorkerPool_Run_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("SMembers error", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		pool := NewWorkerPool(db, nil, nil, 2, 1*time.Second)
		mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetErr(assert.AnError)
		err := pool.Run(ctx)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Empty serverIDs", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		pool := NewWorkerPool(db, nil, nil, 2, 1*time.Second)
		mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetVal([]string{})
		err := pool.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("HGet error", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		monService := new(mockMonitoringService)
		pinger := new(mockPinger)
		pool := NewWorkerPool(db, monService, pinger, 2, 1*time.Second)

		mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetVal([]string{"id-1"})
		mockRedis.ExpectHGet("server:info:id-1", "ipv4").SetErr(assert.AnError)

		err := pool.Run(ctx)
		assert.NoError(t, err) // Doesn't fail overall run
	})

	t.Run("Empty IPv4", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		monService := new(mockMonitoringService)
		pinger := new(mockPinger)
		pool := NewWorkerPool(db, monService, pinger, 2, 1*time.Second)

		mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetVal([]string{"id-1"})
		mockRedis.ExpectHGet("server:info:id-1", "ipv4").SetVal("")

		err := pool.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("Evaluate error", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		monService := new(mockMonitoringService)
		pinger := new(mockPinger)
		pool := NewWorkerPool(db, monService, pinger, 2, 1*time.Second)

		mockRedis.ExpectSMembers(infraRedis.ServerAllIDsKey).SetVal([]string{"id-1"})
		mockRedis.ExpectHGet("server:info:id-1", "ipv4").SetVal("1.1.1.1")
		pinger.On("Ping", "1.1.1.1", 1*time.Second).Return(true)
		monService.On("Evaluate", mock.Anything, "id-1", "1.1.1.1", true).Return(assert.AnError)

		err := pool.Run(ctx)
		assert.NoError(t, err)
	})
}
