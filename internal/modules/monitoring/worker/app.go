package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"server-management-service/internal/modules/monitoring/repository/impl"
	"server-management-service/internal/modules/monitoring/service"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"

	"github.com/redis/go-redis/v9"
)

type App struct {
	RedisClient redis.UniversalClient
	Pool        Pool
}

func NewApp() (*App, error) {
	// Load Configurations
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	}

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		log.Printf("Failed to load redis config: %v", err)
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

	// Initialize Dependencies
	repo := impl.NewGormMonitoringRepository(db)
	txManager := impl.NewGormTxManager(db)
	monService := service.NewMonitoringService(repo, redisClient, txManager)

	// Unprivileged ping for non-root environments (Set to true if running as root on Linux)
	privilegedStr := os.Getenv("ICMP_PRIVILEGED")
	privileged, _ := strconv.ParseBool(privilegedStr)
	pinger := NewICMPPinger(privileged)

	// Settings
	concurrency, _ := config.GetEnvInt("MONITORING_WORKER_CONCURRENCY", 100)
	pingTimeout, _ := config.GetEnvDuration("MONITORING_WORKER_PING_TIMEOUT", 3*time.Second)

	pool := NewWorkerPool(redisClient, monService, pinger, concurrency, pingTimeout)

	return &App{
		RedisClient: redisClient,
		Pool:        pool,
	}, nil
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

	log.Printf("Monitoring Worker started. Scanning every %s\n", tickInterval)

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
	log.Println("Shutting down Monitoring Worker...")
	cancel()

	// Wait a bit for running workers to finish
	time.Sleep(2 * time.Second)
	log.Println("Monitoring Worker stopped.")

	return nil
}

func (a *App) runCycle(ctx context.Context) {
	lockKey := config.GetEnvDefault("MONITORING_WORKER_LOCK_KEY", "lock:monitoring_worker")
	lockExpiration, _ := config.GetEnvDuration("MONITORING_WORKER_LOCK_EXPIRATION", 25*time.Second)

	// Try to acquire distributed lock
	acquired, err := database.AcquireLock(ctx, a.RedisClient, lockKey, lockExpiration)
	if err != nil {
		log.Printf("[Scheduler] Failed to acquire lock: %v\n", err)
		return
	}

	if !acquired {
		log.Println("[Scheduler] Lock not acquired. Another instance is running the cycle.")
		return
	}
	// Not release to ensure the full 25s window belongs to this instance.
	log.Println("[Scheduler] Lock acquired. Starting monitoring cycle...")

	start := time.Now()
	err = a.Pool.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[Scheduler] Cycle completed with error: %v (Duration: %s)\n", err, duration)
	} else {
		log.Printf("[Scheduler] Cycle completed successfully (Duration: %s)\n", duration)
	}
}
