package repository

import (
	"context"
	"time"

	"pinterest-core/internal/domain"
)

type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID int64) error
	Unfollow(ctx context.Context, followerID, followingID int64) error
	ListFollowers(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error)
	ListFollowing(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error)
}
