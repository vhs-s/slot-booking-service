package user

import (
	"context"
	"slot-booking-service-api/internal/repositories/postgres/user"
)

type UserRepo interface {
	Create(ctx context.Context, user *user.User) error
	GetById(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}

type TokenRepo interface {
	SetCachedToken(ctx context.Context, userId, token string) error
}
