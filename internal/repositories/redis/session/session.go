package session

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type repository struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *repository {
	return &repository{rdb: rdb}
}

func (r *repository) SetCachedToken(ctx context.Context, userId, token string) error {
	key := fmt.Sprint("UserSession:" + userId)
	fmt.Println(key)
	err := r.rdb.HSet(ctx, key, map[string]interface{}{
		"token": token,
	}).Err()

	if err != nil {
		log.Println("unable to set userSession:", err, userId, token)
		return err
	}
	return nil
}

func (r *repository) GetCachedToken(ctx context.Context, userId string) (string, error) {
	key := fmt.Sprint("UserSession:" + userId)
	cachedToken, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		log.Println("unable to get userSession:", err, userId, cachedToken)
		return "", err
	}
	return cachedToken["token"], nil
}
