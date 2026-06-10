package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/modules/monitoring/repository"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"github.com/redis/go-redis/v9"
)

type MonitoringService interface {
	Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error
}

type monitoringServiceImpl struct {
	repo             repository.MonitoringRepository
	rdb              redis.UniversalClient
	esLogger         elasticsearch.ObservationLogger
	failureThreshold int
}

func NewMonitoringService(repo repository.MonitoringRepository, rdb redis.UniversalClient, esLogger elasticsearch.ObservationLogger, failureThreshold int) MonitoringService {
	return &monitoringServiceImpl{
		repo:             repo,
		rdb:              rdb,
		esLogger:         esLogger,
		failureThreshold: failureThreshold,
	}
}

func (s *monitoringServiceImpl) Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error {
	// 1. Log observation directly to Elasticsearch (Time-Series)
	// We MUST log every single ping attempt to ES to calculate Uptime later
	err := s.esLogger.LogObservation(ctx, serverID, pingSuccess)
	if err != nil {
		// Log the error but don't fail the rest of the evaluation
		log.Printf("[WARNING] failed to log observation to ES for server %s: %v", serverID, err)
	}

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

	if pingSuccess {
		if currentStatus == serverDomain.ServerStatusOffline {
			// Recovery Threshold = 1
			newStatus = serverDomain.ServerStatusOnline
			statusChanged = true
			retryCount = 0
		} else {
			// Already online, reset retry count if > 0
			retryCount = 0
		}
	} else {
		if currentStatus == serverDomain.ServerStatusOnline {
			retryCount++
			if retryCount >= s.failureThreshold {
				newStatus = serverDomain.ServerStatusOffline
				statusChanged = true
				retryCount = 0
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

	// Update Postgres ONLY if state actually changes
	if statusChanged {
		err = s.repo.UpdateServerStatus(ctx, serverID, newStatus, retryCount)
		if err != nil {
			return fmt.Errorf("failed to update server status in postgres: %w", err)
		}
	} else if retryCount > 0 && !pingSuccess && currentStatus == serverDomain.ServerStatusOnline {
		// Update consecutive failures even if status hasn't changed to offline yet
		err = s.repo.UpdateServerStatus(ctx, serverID, currentStatus, retryCount)
		if err != nil {
			return fmt.Errorf("failed to update consecutive failures in postgres: %w", err)
		}
	}

	return nil
}
