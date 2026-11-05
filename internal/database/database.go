package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func ConnectPostgres(dsn string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Print("Подключено PostgreSQL")
	return dbpool
}

func ConnectRedis(dsn string) *redis.Client {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		log.Fatal("Ошибка разбора Redis DSN:", err)
	}
	c := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = c.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Ошибка подключения к Redis:", err)
		return nil
	}
	log.Print("Подключено Redis")
	return c
}
