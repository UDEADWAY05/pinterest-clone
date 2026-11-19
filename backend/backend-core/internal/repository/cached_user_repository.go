package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pinterest-core/internal/domain"
)

type CachedUserRepository struct {
	wrapped UserRepository
	cache   CacheRepository
	ttl     time.Duration
}

func NewCachedUserRepository(
	wrapped UserRepository,
	cache CacheRepository,
	ttl time.Duration,
) *CachedUserRepository {
	return &CachedUserRepository{
		wrapped: wrapped,
		cache:   cache,
		ttl:     ttl,
	}
}

func userByIDKey(id int64) string    {
	return fmt.Sprintf("user:id:%d", id)
}
func userByEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}
func userListKey() string {
	return "user:list:all"
}

func (r *CachedUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	key := userByIDKey(id)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user:id:%d failed: %v", id, err)
	}
	if cached != nil {
		var user domain.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
		log.Printf("cache unmarshal user:id:%d failed: %v", id, err)
	}

	user, err := r.wrapped.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, user)
	return user, nil
}

func (r *CachedUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	key := userByEmailKey(email)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user:email:%s failed: %v", email, err)
	}
	if cached != nil {
		var user domain.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
		log.Printf("cache unmarshal user:email:%s failed: %v", email, err)
	}

	user, err := r.wrapped.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, user)
	r.cacheSet(ctx, userByIDKey(user.ID), user)

	return user, nil
}

func (r *CachedUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	key := userListKey()

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user list failed: %v", err)
	}
	if cached != nil {
		var users []domain.User
		if err := json.Unmarshal(cached, &users); err == nil {
			return users, nil
		}
		log.Printf("cache unmarshal user list failed: %v", err)
	}

	users, err := r.wrapped.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, users)
	return users, nil
}

func (r *CachedUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if err := r.wrapped.CreateUser(ctx, user); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, userListKey()); err != nil {
		log.Printf("cache invalidate user list failed: %v", err)
	}

	return nil
}

func (r *CachedUserRepository) GetPinsCount(ctx context.Context, userID int64) (int, error) {
	return r.wrapped.GetPinsCount(ctx, userID)
}

func (r *CachedUserRepository) GetFollowersCount(ctx context.Context, userID int64) (int, error) {
	return r.wrapped.GetFollowersCount(ctx, userID)
}

func (r *CachedUserRepository) GetFollowingCount(ctx context.Context, userID int64) (int, error) {
	return r.wrapped.GetFollowingCount(ctx, userID)
}

func (r *CachedUserRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return r.wrapped.IsFollowing(ctx, followerID, followingID)
}

func (r *CachedUserRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	return r.wrapped.Follow(ctx, followerID, followingID)
}

func (r *CachedUserRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return r.wrapped.Unfollow(ctx, followerID, followingID)
}

func (r *CachedUserRepository) cacheSet(ctx context.Context, key string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal %s failed: %v", key, err)
		return
	}
	if err := r.cache.Set(ctx, key, data, r.ttl); err != nil {
		log.Printf("cache set %s failed: %v", key, err)
	}
}
