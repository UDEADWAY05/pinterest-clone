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


	// Redis
	RedisAddr string

	// S3-compatible storage
	S3Endpoint      string
	S3Bucket        string
	S3AccessKey     string
	S3SecretKey     string
	S3Region        string
	S3UseSSL        bool
	S3PublicURL     string // публичный базовый URL (если отличается от эндпоинта, например CDN)
	S3PresignExpiry int    // время жизни presigned URL в секундах (по умолчанию 900 = 15 мин)

	// Cache TTL settings (in seconds)
	CachePinTTL   int // TTL для одного пина (по умолчанию 300 = 5 мин)
	CacheListTTL  int // TTL для списков/лент (по умолчанию 60 = 1 мин)
	CacheUserTTL  int // TTL для профиля пользователя (по умолчанию 600 = 10 мин)
	CacheBoardTTL int // TTL для одной доски (по умолчанию 300 = 5 мин)

	//CORS
	AllowedMethods string
	AllowedHeaders string
	AllowedCredentials bool
	AllowedOrigins map[string]bool
}

var allowOrigins map[string]bool = map[string]bool{
	"http://localhost:3000": true,
}

func Load() (*Config, error) {
	c := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		PostgresDSN:        getEnv("POSTGRES_DSN", ""),
		PostgresMaxConns:   int32(getEnvInt("POSTGRES_MAX_CONNS", 10)),
		PostgresMinConns:   int32(getEnvInt("POSTGRES_MIN_CONNS", 5)),
		PostgresTimeoutSec: getEnvInt("POSTGRES_TIMEOUT_SEC", 10),
		RedisAddr:          getEnv("REDIS_ADDR", ""),
		S3Endpoint:         getEnv("S3_ENDPOINT", ""),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		S3AccessKey:        getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:        getEnv("S3_SECRET_KEY", ""),
		S3Region:           getEnv("S3_REGION", "us-east-1"),
		S3UseSSL:           getEnvBool("S3_USE_SSL", false),
		S3PresignExpiry:    getEnvInt("S3_PRESIGN_EXPIRY", 900),
		CachePinTTL:        getEnvInt("CACHE_PIN_TTL", 300),
		CacheListTTL:       getEnvInt("CACHE_LIST_TTL", 60),
		CacheUserTTL:       getEnvInt("CACHE_USER_TTL", 600),
		CacheBoardTTL:      getEnvInt("CACHE_BOARD_TTL", 300),

		AllowedMethods: "GET, POST, PUT, DELETE, OPTIONS",
		AllowedHeaders: "Content-Type, Authorization",
		AllowedCredentials: true,
		AllowedOrigins: allowOrigins,
	}

	if c.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}

	return c, nil
}

func (c *Config) PostgresTimeout() time.Duration {
	return time.Duration(c.PostgresTimeoutSec) * time.Second
}

// PresignExpiry возвращает длительность жизни presigned URL.
func (c *Config) PresignExpiry() time.Duration {
	return time.Duration(c.S3PresignExpiry) * time.Second
}

// CacheTTLs возвращает TTL для разных типов кешируемых данных.
func (c *Config) PinCacheTTL() time.Duration   { return time.Duration(c.CachePinTTL) * time.Second }
func (c *Config) ListCacheTTL() time.Duration  { return time.Duration(c.CacheListTTL) * time.Second }
func (c *Config) UserCacheTTL() time.Duration  { return time.Duration(c.CacheUserTTL) * time.Second }
func (c *Config) BoardCacheTTL() time.Duration { return time.Duration(c.CacheBoardTTL) * time.Second }

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

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
