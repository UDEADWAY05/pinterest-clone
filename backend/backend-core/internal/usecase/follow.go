package usecase

import (
	"context"
	"errors"
	"time"

	"pinterest-core/internal/domain"
	repository "pinterest-core/internal/repository"
)

type FollowUsecase struct {
	Repo repository.FollowRepository
}

func NewFollowUsecase(repo repository.FollowRepository) *FollowUsecase {
	return &FollowUsecase{Repo: repo}
}

func (u *FollowUsecase) Follow(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}
	return u.Repo.Follow(ctx, followerID, followingID)
}

func (u *FollowUsecase) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return u.Repo.Unfollow(ctx, followerID, followingID)
}

func (u *FollowUsecase) ListFollowers(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	return u.Repo.ListFollowers(ctx, userID, cursor, limit)
}

func (u *FollowUsecase) ListFollowing(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	return u.Repo.ListFollowing(ctx, userID, cursor, limit)
}
