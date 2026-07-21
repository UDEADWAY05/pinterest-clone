package repository

import (
	"context"
	"time"
)


type StorageRepository interface {
	PresignPutURL(ctx context.Context, key string, expiry time.Duration) (uploadURL string, err error)
	PublicURL(key string) string
}
