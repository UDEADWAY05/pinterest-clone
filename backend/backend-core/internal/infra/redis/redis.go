package redis

import (
	"context"
	"fmt"
	"time"

	"pinterest-core/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *goredis.Client
}

func New(ctx context.Context, cfg *config.Config) (*Redis, error) {
	opts := &goredis.Options{
		Addr:         cfg.RedisAddr,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Close(context.Context) error {
	if r == nil || r.Client == nil {
		return nil
	}

	return r.Client.Close()
}
