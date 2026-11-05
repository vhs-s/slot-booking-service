package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slot-booking-service-api/internal/auth"

	"github.com/redis/go-redis/v9"
)

type repository struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *repository {
	return &repository{rdb: rdb}
}

func (r *repository) SetCachedToken(ctx context.Context, userId, sessionId, token string) error {
	key := sessionKey(userId, sessionId)
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, key, "token", token)
	pipe.Expire(ctx, key, auth.RefreshTokenTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Println("unable to set user session:", err, userId, sessionId)
		return err
	}
	return nil
}

func (r *repository) GetCachedToken(ctx context.Context, userId, sessionId string) (string, error) {
	cachedToken, err := r.rdb.HGet(ctx, sessionKey(userId, sessionId), "token").Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		log.Println("unable to get user session:", err, userId, sessionId)
		return "", err
	}
	return cachedToken, nil
}

func (r *repository) DeleteCachedToken(ctx context.Context, userId, sessionId string) error {
	if err := r.rdb.Del(ctx, sessionKey(userId, sessionId)).Err(); err != nil {
		log.Println("unable to delete user session:", err, userId, sessionId)
		return err
	}
	return nil
}

func sessionKey(userId, sessionId string) string {
	return fmt.Sprintf("UserSession:%s:%s", userId, sessionId)
}
