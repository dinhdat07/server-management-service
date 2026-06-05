package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	infraRedis "server-management-service/internal/infrastructure/redis"
	"server-management-service/internal/modules/monitoring/service"

	"github.com/redis/go-redis/v9"
)

type Pool interface {
	Run(ctx context.Context) error
}

type workerPool struct {
	rdb         redis.UniversalClient
	monService  service.MonitoringService
	pinger      Pinger
	concurrency int
	timeout     time.Duration
}

func NewWorkerPool(rdb redis.UniversalClient, monService service.MonitoringService, pinger Pinger, concurrency int, timeout time.Duration) Pool {
	return &workerPool{
		rdb:         rdb,
		monService:  monService,
		pinger:      pinger,
		concurrency: concurrency,
		timeout:     timeout,
	}
}

func (w *workerPool) Run(ctx context.Context) error {
	// Fetch all Server IDs from Redis
	serverIDs, err := w.rdb.SMembers(ctx, infraRedis.ServerAllIDsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch server IDs from redis: %w", err)
	}

	if len(serverIDs) == 0 {
		return nil // Nothing to do
	}

	// Setup Worker Pool
	var wg sync.WaitGroup
	idChan := make(chan string, len(serverIDs))

	// Spawn workers
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idChan {
				w.processServer(ctx, id)
			}
		}()
	}

	// Distribute work
	for _, id := range serverIDs {
		idChan <- id
	}
	close(idChan)

	// Wait for all workers to finish
	wg.Wait()
	return nil
}

func (w *workerPool) processServer(ctx context.Context, serverID string) {
	// Get server info (IPv4) from Redis
	redisKey := fmt.Sprintf(infraRedis.ServerInfoKeyFmt, serverID)
	ipv4, err := w.rdb.HGet(ctx, redisKey, "ipv4").Result()
	if err != nil {
		log.Printf("[Worker] Failed to get IP for server %s: %v\n", serverID, err)
		return
	}
	if ipv4 == "" {
		log.Printf("[Worker] IPv4 is empty for server %s\n", serverID)
		return
	}

	// Perform Ping
	success := w.pinger.Ping(ipv4, w.timeout)

	// Evaluate State Machine
	err = w.monService.Evaluate(ctx, serverID, ipv4, success)
	if err != nil {
		log.Printf("[Worker] Failed to evaluate state for server %s (IP: %s): %v\n", serverID, ipv4, err)
	}
}
