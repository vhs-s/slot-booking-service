package auth

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=contracts.go -destination=mocks/token_repo.go -package=mocks
type TokenRepo interface { //rdb
	SetCachedToken(ctx context.Context, userId, token string) error
	GetCachedToken(ctx context.Context, userId string) (string, error)
}
