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

func ConnectPostgres() *pgxpool.Pool {
	dsn := "postgresql://postgres:root@localhost:5432/booking_db?sslmode=disable"

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Print("Подключено PostgreSQL")
	return dbpool
}

func ConnectRedis() *redis.Client {
	c := redis.NewClient(
		&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := c.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Ошибка подключения к Redis:", err)
		return nil
	}
	log.Print("Подключено Redis")
	return c
}
