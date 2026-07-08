package repository

import (
	"context"
	"time"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error
	GetUserIDByTokenHash(ctx context.Context, tokenHash string) (int64, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
}
