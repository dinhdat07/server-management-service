package main

import (
	"context"
	"fmt"
	"os"
	"server-management-service/internal/shared/logger"
	"strconv"
	"time"

	infraRedis "server-management-service/internal/infrastructure/redis"
	"server-management-service/internal/modules/server_management/domain"
	repimpl "server-management-service/internal/modules/server_management/repository/impl"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"

	"github.com/google/uuid"
)

func main() {
	if err := run(); err != nil {
		logger.Log.Sugar().Fatalf("seed failed: %v", err)
	}
}

func run() error {
	count := 10000
	if v := os.Getenv("SIMULATOR_IP_COUNT"); v != "" {
		count, _ = strconv.Atoi(v)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.GetInstance(cfg.DBUrl)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}

	// Clean up previous simulation data
	logger.Log.Sugar().Info("Cleaning up previous simulation servers...")
	result := db.Where("server_name LIKE ?", "sim-%").Delete(&domain.Server{})
	if result.Error != nil {
		return fmt.Errorf("cleanup old sim servers: %w", result.Error)
	}
	logger.Log.Sugar().Infof("Deleted %d old simulation servers", result.RowsAffected)

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		return fmt.Errorf("redis config: %w", err)
	}
	redisClient := database.NewRedisClient(redisCfg)
	cache := infraRedis.NewServerCache(redisClient)

	// Clean Redis: flush all server data (will be repopulated by seed)
	if redisClient != nil {
		ctx := context.Background()
		keys, _ := redisClient.Keys(ctx, "server:*").Result()
		if len(keys) > 0 {
			redisClient.Del(ctx, keys...)
		}
		logger.Log.Sugar().Infof("Redis: flushed %d server keys", len(keys))
	}

	repo := repimpl.NewGormServerRepository(db)

	batchSize := 500
	subnet := "10.1"
	octet3 := 0
	octet4 := 1

	logger.Log.Sugar().Infof("Seeding %d servers in batches of %d...", count, batchSize)

	for i := 0; i < count; i += batchSize {
		end := i + batchSize
		if end > count {
			end = count
		}
		batch := make([]*domain.Server, 0, end-i)
		cacheItems := make([]infraRedis.CacheUpsertItem, 0, end-i)

		for j := i; j < end; j++ {
			ip := fmt.Sprintf("%s.%d.%d", subnet, octet3, octet4)
			id := uuid.New().String()
			name := fmt.Sprintf("sim-%s", id[:8])

			batch = append(batch, &domain.Server{
				ServerID:      id,
				ServerName:    name,
				IPv4:          ip,
				CurrentStatus: domain.ServerStatusOnline,
			})

			cacheItems = append(cacheItems, infraRedis.CacheUpsertItem{
				ID:         id,
				IPv4:       ip,
				Status:     "ONLINE",
				RetryCount: 0,
			})

			octet4++
			if octet4 > 254 {
				octet4 = 1
				octet3++
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := repo.BatchCreate(ctx, batch); err != nil {
			cancel()
			return fmt.Errorf("batch create at offset %d: %w", i, err)
		}
		cancel()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		if err := cache.BatchUpsert(ctx2, cacheItems); err != nil {
			cancel2()
			logger.Log.Sugar().Warnf("WARN: redis batch upsert at offset %d: %v", i, err)
		}
		cancel2()

		logger.Log.Sugar().Infof("Seeded %d/%d servers", end, count)
	}

	logger.Log.Sugar().Infof("Seed complete: %d servers in DB + Redis", count)
	return nil
}
