package service

import (
	"context"
	"fmt"
	"strconv"

	"server-management-service/internal/modules/monitoring/domain"
	"server-management-service/internal/modules/monitoring/repository"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"github.com/redis/go-redis/v9"
)

type MonitoringService interface {
	Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error
}

type monitoringServiceImpl struct {
	repo      repository.MonitoringRepository
	rdb       redis.UniversalClient
	txManager repository.TxManager
}

func NewMonitoringService(repo repository.MonitoringRepository, rdb redis.UniversalClient, txManager repository.TxManager) MonitoringService {
	return &monitoringServiceImpl{
		repo:      repo,
		rdb:       rdb,
		txManager: txManager,
	}
}

func (s *monitoringServiceImpl) Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error {
	redisKey := fmt.Sprintf("server:info:%s", serverID)

	// Fetch current status and retry count from Redis
	vals, err := s.rdb.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get server info from redis: %w", err)
	}

	if len(vals) == 0 {
		return fmt.Errorf("server info not found in redis for id: %s", serverID)
	}

	currentStatusStr := vals["status"]
	if currentStatusStr == "" {
		currentStatusStr = string(serverDomain.ServerStatusOnline) // Default
	}
	currentStatus := serverDomain.ServerStatus(currentStatusStr)

	retryCountStr := vals["retry_count"]
	retryCount := 0
	if retryCountStr != "" {
		retryCount, _ = strconv.Atoi(retryCountStr)
	}

	// State Machine Evaluation
	var newStatus serverDomain.ServerStatus
	var statusChanged bool
	var reason string

	if pingSuccess {
		if currentStatus == serverDomain.ServerStatusOffline {
			// Recovery Threshold = 1
			newStatus = serverDomain.ServerStatusOnline
			statusChanged = true
			retryCount = 0
			reason = "ICMP Ping Succeeded (Recovery)"
		} else {
			// Already online, reset retry count if > 0
			retryCount = 0
		}
	} else {
		if currentStatus == serverDomain.ServerStatusOnline {
			retryCount++
			if retryCount >= 2 { // Failure Threshold = 2
				newStatus = serverDomain.ServerStatusOffline
				statusChanged = true
				retryCount = 0
				reason = "ICMP Ping Failed 2 consecutive times"
			}
		} else {
			// Already offline
			retryCount++
		}
	}

	// Update Redis cache
	if statusChanged {
		err = s.rdb.HSet(ctx, redisKey, "status", string(newStatus), "retry_count", retryCount).Err()
	} else {
		err = s.rdb.HSet(ctx, redisKey, "status", string(currentStatus), "retry_count", retryCount).Err()
	}
	if err != nil {
		return fmt.Errorf("failed to update redis status: %w", err)
	}

	// Persist the transition and update Postgres
	if statusChanged {
		event := &domain.StatusTransitionEvent{
			ServerID:       serverID,
			PreviousStatus: string(currentStatus),
			CurrentStatus:  string(newStatus),
			Reason:         reason,
		}

		err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
			return s.repo.SaveTransitionAndUpdateServer(txCtx, event, newStatus)
		})
		if err != nil {
			return fmt.Errorf("failed to save transition to db: %w", err)
		}
	}

	return nil
}
