package repository

import (
	"context"
	"time"

	"pinterest-core/internal/domain"
)

type CachedFollowRepository struct {
	wrapped FollowRepository
}

func NewCachedFollowRepository(wrapped FollowRepository) *CachedFollowRepository {
	return &CachedFollowRepository{wrapped: wrapped}
}

func (r *CachedFollowRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	return r.wrapped.Follow(ctx, followerID, followingID)
}

func (r *CachedFollowRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return r.wrapped.Unfollow(ctx, followerID, followingID)
}

func (r *CachedFollowRepository) ListFollowers(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	return r.wrapped.ListFollowers(ctx, userID, cursor, limit)
}

func (r *CachedFollowRepository) ListFollowing(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	return r.wrapped.ListFollowing(ctx, userID, cursor, limit)
}
