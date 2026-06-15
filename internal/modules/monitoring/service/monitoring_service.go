package service

import (
	"context"
	"fmt"

	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/modules/monitoring/repository"
	serverDomain "server-management-service/internal/modules/server_management/domain"
)

type MonitoringService interface {
	Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error
}

type monitoringServiceImpl struct {
	repo             repository.MonitoringRepository
	stateStore       repository.ServerStateStore
	esLogger         elasticsearch.ObservationLogger
	failureThreshold int
}

func NewMonitoringService(repo repository.MonitoringRepository, stateStore repository.ServerStateStore, esLogger elasticsearch.ObservationLogger, failureThreshold int) MonitoringService {
	return &monitoringServiceImpl{
		repo:             repo,
		stateStore:       stateStore,
		esLogger:         esLogger,
		failureThreshold: failureThreshold,
	}
}

func (s *monitoringServiceImpl) Evaluate(ctx context.Context, serverID string, ip string, pingSuccess bool) error {
	// Fire-and-forget: buffered, non-blocking, flushed in bulk
	s.esLogger.LogObservation(ctx, serverID, pingSuccess)

	// Fetch current status and retry count from state store (Redis)
	state, err := s.stateStore.GetServerState(ctx, serverID)
	if err != nil {
		return err
	}

	currentStatusStr := state.Status
	if currentStatusStr == "" {
		currentStatusStr = string(serverDomain.ServerStatusOnline) // Default
	}
	currentStatus := serverDomain.ServerStatus(currentStatusStr)

	retryCount := state.RetryCount

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
		err = s.stateStore.SetServerState(ctx, serverID, string(newStatus), retryCount)
	} else {
		err = s.stateStore.SetServerState(ctx, serverID, string(currentStatus), retryCount)
	}
	if err != nil {
		return err
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
