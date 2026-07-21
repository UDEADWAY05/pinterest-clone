package repository

import (
	"context"
	"time"
)


type CacheRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
}
