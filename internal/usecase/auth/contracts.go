package auth

import "context"

type TokenRepo interface { //rdb
	SetCachedToken(ctx context.Context, userId, token string) error
	GetCachedToken(ctx context.Context, userId string) (string, error)
}
