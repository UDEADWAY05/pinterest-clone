package postgresrepo

import (
	"context"
	"fmt"
	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}



func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{Pool: pool}
}


func (r *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users
		ORDER BY created_at DESC`)

	if err != nil {
		fmt.Print(err)
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.CreatedAt); err != nil {
			fmt.Print(err)
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User

	err := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users WHERE id = $1`,
		id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, mapNoRows(err, repository.ErrUserNotFound)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, created_at
		FROM users
		WHERE email = $1`,
		email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, mapNoRows(err, repository.ErrUserNotFound)
	}

	return &user, nil
}

func mapNoRows(err error, notFound error) error {
	if err.Error() == "no rows in result set" {
		return notFound
	}
	return fmt.Errorf("database query failed: %w", err)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		user.Email, user.PasswordHash, user.DisplayName, user.AvatarURL,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetPinsCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pins
		WHERE author_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query pins count: %w", err)
	}
	return count, nil
}

func (r *UserRepository) GetFollowersCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM follows WHERE following_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query followers count: %w", err)
	}
	return count, nil
}

func (r *UserRepository) GetFollowingCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM follows WHERE follower_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query following count: %w", err)
	}
	return count, nil
}

func (r *UserRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	var exists bool
	err := r.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query is following: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
		followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("follow: %w", err)
	}
	return nil
}

func (r *UserRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	_, err := r.Pool.Exec(ctx, `
		DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	return nil
}
