package monitoring

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/modules/monitoring/repository/impl"
	"server-management-service/internal/modules/monitoring/service"
	"server-management-service/internal/modules/monitoring/worker"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
	"server-management-service/internal/shared/logger"

	"github.com/redis/go-redis/v9"
)

type App struct {
	RedisClient redis.UniversalClient
	Pool        worker.Pool
	esLogger    elasticsearch.ObservationLogger
}

func NewApp() (*App, error) {
	// Initialize logger
	cfg, _ := config.Load()
	if cfg != nil {
		logger.InitLogger(cfg.Logger, "monitoring-worker")
	} else {
		logger.InitLogger(config.LoggerConfig{}, "monitoring-worker")
	}

	// Load Configurations
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	}

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to load redis config: %v", err)
	}

	// Initialize Postgres
	db, err := database.GetInstance(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize Redis
	redisClient := database.NewRedisClient(redisCfg)
	if redisClient == nil {
		return nil, fmt.Errorf("redis is required for Monitoring Worker")
	}
	if err := database.PingRedis(context.Background(), redisClient); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	// Initialize Elasticsearch
	esCfg := config.LoadElasticsearchConfig()
	esClient, err := database.NewElasticsearchClient(context.Background(), []string{esCfg.URL})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connection failed: %w", err)
	}
	esLogger := elasticsearch.NewObservationLogger(esClient, esCfg.ServerIndex, config.LoadObservationLoggerConfig())

	// Initialize Dependencies
	repo := impl.NewGormMonitoringRepository(db)
	stateStore := impl.NewRedisServerStateStore(redisClient)
	threshold, _ := config.GetEnvInt("MONITORING_FAILURE_THRESHOLD", 2)
	monService := service.NewMonitoringService(repo, stateStore, esLogger, threshold)

	// Unprivileged ping for non-root environments (Set to true if running as root on Linux)
	privilegedStr := os.Getenv("ICMP_PRIVILEGED")
	privileged, _ := strconv.ParseBool(privilegedStr)
	pinger := worker.NewICMPPinger(privileged)

	// Settings
	concurrency, _ := config.GetEnvInt("MONITORING_WORKER_CONCURRENCY", 100)
	pingTimeout, _ := config.GetEnvDuration("MONITORING_WORKER_PING_TIMEOUT", 3*time.Second)

	pool := worker.NewWorkerPool(redisClient, monService, pinger, concurrency, pingTimeout)

	return &App{
		RedisClient: redisClient,
		Pool:        pool,
		esLogger:    esLogger,
	}, nil
}

func (a *App) Shutdown() {
	if a.esLogger != nil {
		a.esLogger.Shutdown()
	}
}

func (a *App) Run() error {
	tickInterval, _ := config.GetEnvDuration("MONITORING_WORKER_TICK_INTERVAL", 30*time.Second)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)

	// Start Scheduler
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	logger.Log.Sugar().Infof("Monitoring Worker started. Scanning every %s\n", tickInterval)

	go func() {
		// Run immediately on startup
		a.runCycle(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.runCycle(ctx)
			}
		}
	}()

	<-sigCh
	logger.Log.Sugar().Info("Shutting down Monitoring Worker...")
	cancel()

	// Wait for running workers to finish
	time.Sleep(2 * time.Second)

	if a.esLogger != nil {
		a.esLogger.Shutdown()
	}

	logger.Log.Sugar().Info("Monitoring Worker stopped.")
	return nil
}

func (a *App) runCycle(ctx context.Context) {
	lockKey := config.GetEnvDefault("MONITORING_PRODUCER_LOCK_KEY", "lock:monitoring_producer")
	lockExpiration := 10 * time.Second

	// 1. PHASE 1: Producer Election
	acquired, _ := database.AcquireLock(ctx, a.RedisClient, lockKey, lockExpiration)
	if acquired {
		logger.Log.Sugar().Info("[Producer] Lock acquired. Populating work queue...")
		// Fetch all Server IDs
		serverIDs, err := a.RedisClient.SMembers(ctx, "server:all_ids").Result()
		if err == nil && len(serverIDs) > 0 {
			// Clear existing queue just in case
			a.RedisClient.Del(ctx, "monitoring:queue")
			
			// Push all servers to queue
			args := make([]interface{}, len(serverIDs))
			for i, v := range serverIDs {
				args[i] = v
			}
			a.RedisClient.RPush(ctx, "monitoring:queue", args...)
			logger.Log.Sugar().Infof("[Producer] Pushed %d servers to the queue.", len(serverIDs))
		}
	} else {
		logger.Log.Sugar().Info("[Consumer] Ready to process queue...")
		// Small wait to allow Producer to initialize queue
		time.Sleep(1 * time.Second)
	}

	// 2. PHASE 2: Consumer (ALL workers participate)
	start := time.Now()
	err := a.Pool.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.Log.Sugar().Errorf("[Consumer] Worker cycle completed with error: %v (Duration: %s)\n", err, duration)
	} else {
		logger.Log.Sugar().Infof("[Consumer] Worker cycle completed (Duration: %s)\n", duration)
	}
}
