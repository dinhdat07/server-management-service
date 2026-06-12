package repository

import (
	"context"

	"server-management-service/internal/modules/identity/domain"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uint) (*domain.User, error)
}
