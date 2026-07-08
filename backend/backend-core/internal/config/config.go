package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	ServerPort string
	LogLevel   string

	// PostgreSQL
	PostgresDSN        string
	PostgresMaxConns   int32
	PostgresMinConns   int32
	PostgresTimeoutSec int

	// Auth
	JWTSecret string

	// Optional infra
	RedisAddr  string
	S3Endpoint string
	S3Bucket   string
}

func Load() (*Config, error) {
	c := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		PostgresDSN:        getEnv("POSTGRES_DSN", ""),
		PostgresMaxConns:   int32(getEnvInt("POSTGRES_MAX_CONNS", 10)),
		PostgresMinConns:   int32(getEnvInt("POSTGRES_MIN_CONNS", 5)),
		PostgresTimeoutSec: getEnvInt("POSTGRES_TIMEOUT_SEC", 10),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		RedisAddr:          getEnv("REDIS_ADDR", ""),
		S3Endpoint:         getEnv("S3_ENDPOINT", ""),
		S3Bucket:           getEnv("S3_BUCKET", ""),
	}

	if c.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}

	return c, nil
}

func (c *Config) PostgresTimeout() time.Duration {
	return time.Duration(c.PostgresTimeoutSec) * time.Second
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
