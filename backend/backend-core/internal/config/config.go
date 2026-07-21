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

	// Cache TTL settings (in seconds)
	CachePinTTL  int // TTL для одного пина (по умолчанию 300 = 5 мин)
	CacheListTTL int // TTL для списков/лент (по умолчанию 60 = 1 мин)
	CacheUserTTL int // TTL для профиля пользователя (по умолчанию 600 = 10 мин)
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
		CachePinTTL:        getEnvInt("CACHE_PIN_TTL", 300),
		CacheListTTL:       getEnvInt("CACHE_LIST_TTL", 60),
		CacheUserTTL:       getEnvInt("CACHE_USER_TTL", 600),
	}

	if c.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}

	return c, nil
}

func (c *Config) PostgresTimeout() time.Duration {
	return time.Duration(c.PostgresTimeoutSec) * time.Second
}

// CacheTTLs возвращает TTL для разных типов кешируемых данных.
// Вынесено в методы, чтобы не дублировать time.Duration в месте использования.
func (c *Config) PinCacheTTL() time.Duration   { return time.Duration(c.CachePinTTL) * time.Second }
func (c *Config) ListCacheTTL() time.Duration  { return time.Duration(c.CacheListTTL) * time.Second }
func (c *Config) UserCacheTTL() time.Duration  { return time.Duration(c.CacheUserTTL) * time.Second }

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
