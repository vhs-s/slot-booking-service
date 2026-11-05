package auth

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=contracts.go -destination=mocks/token_repo.go -package=mocks
type TokenRepo interface {
	SetCachedToken(ctx context.Context, userId, sessionId, token string) error
	GetCachedToken(ctx context.Context, userId, sessionId string) (string, error)
	DeleteCachedToken(ctx context.Context, userId, sessionId string) error
}
