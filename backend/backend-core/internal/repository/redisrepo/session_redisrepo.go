package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) CreateSession(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error {
	key := fmt.Sprintf("session:%s", tokenHash)
	ttl := time.Until(expiresAt)

	err := r.client.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetUserIDByTokenHash(ctx context.Context, tokenHash string) (int64, error) {
	key := fmt.Sprintf("session:%s", tokenHash)

	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, fmt.Errorf("session not found")
	}
	if err != nil {
		return 0, fmt.Errorf("redis get session: %w", err)
	}

	return val, nil
}

func (r *SessionRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	key := fmt.Sprintf("session:%s", tokenHash)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete session: %w", err)
	}
	return nil
}
