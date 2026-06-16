package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"server-management-service/internal/modules/identity/repository"
)

type revocationStoreRedis struct {
	client redis.UniversalClient
}

func NewSessionRevocationStore(client redis.UniversalClient) repository.SessionRevocationStore {
	return &revocationStoreRedis{client: client}
}

func (s *revocationStoreRedis) MarkRevoked(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // Already expired
	}

	key := "revoked_session:" + sessionID.String()
	return s.client.Set(ctx, key, "revoked", ttl).Err()
}
