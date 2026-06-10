package impl

import (
	"context"
	"fmt"
	"server-management-service/internal/modules/monitoring/repository"

	"github.com/redis/go-redis/v9"
)

// RedisServerStateStore implements domain.ServerStateStore using Redis hashes.
type RedisServerStateStore struct {
	rdb redis.UniversalClient
}

func NewRedisServerStateStore(rdb redis.UniversalClient) repository.ServerStateStore {
	return &RedisServerStateStore{rdb: rdb}
}

func (s *RedisServerStateStore) GetServerState(ctx context.Context, serverID string) (map[string]string, error) {
	redisKey := fmt.Sprintf("server:info:%s", serverID)
	vals, err := s.rdb.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info from redis: %w", err)
	}
	return vals, nil
}

func (s *RedisServerStateStore) SetServerState(ctx context.Context, serverID string, status string, retryCount int) error {
	redisKey := fmt.Sprintf("server:info:%s", serverID)
	err := s.rdb.HSet(ctx, redisKey, "status", status, "retry_count", retryCount).Err()
	if err != nil {
		return fmt.Errorf("failed to update redis status: %w", err)
	}
	return nil
}



