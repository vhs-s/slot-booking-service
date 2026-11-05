package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	envHTTPPort         = "HTTP_PORT"
	envPostgresDSN      = "POSTGRES_DSN"
	envRedisDSN         = "REDIS_DSN"
	envJWTAccessSecret  = "JWT_ACCESS_SECRET"
	envJWTRefreshSecret = "JWT_REFRESH_SECRET"
)

type Config struct {
	HTTPPort         string
	PostgresDSN      string
	RedisDSN         string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:         os.Getenv(envHTTPPort),
		PostgresDSN:      os.Getenv(envPostgresDSN),
		RedisDSN:         os.Getenv(envRedisDSN),
		JWTAccessSecret:  os.Getenv(envJWTAccessSecret),
		JWTRefreshSecret: os.Getenv(envJWTRefreshSecret),
	}

	requiredValues := []struct {
		name  string
		value string
	}{
		{name: envHTTPPort, value: cfg.HTTPPort},
		{name: envPostgresDSN, value: cfg.PostgresDSN},
		{name: envRedisDSN, value: cfg.RedisDSN},
		{name: envJWTAccessSecret, value: cfg.JWTAccessSecret},
		{name: envJWTRefreshSecret, value: cfg.JWTRefreshSecret},
	}

	for _, required := range requiredValues {
		if strings.TrimSpace(required.value) == "" {
			return Config{}, fmt.Errorf("required environment variable %s is not set", required.name)
		}
	}

	return cfg, nil
}
