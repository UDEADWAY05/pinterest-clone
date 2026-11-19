package postgresrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

type FollowRepository struct {
	Pool *pgxpool.Pool
}

func NewFollowRepository(pool *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{Pool: pool}
}

func (r *FollowRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, following_id) DO NOTHING;`,
		followerID, followingID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return repository.ErrUserNotFound
		}
		return fmt.Errorf("follow: %w", err)
	}
	return nil
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	_, err := r.Pool.Exec(ctx, `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2;`,
		followerID, followingID,
	)

	if err != nil {
		return fmt.Errorf("unfollow: %w", err)
	}
	return nil
}

func (r *FollowRepository) ListFollowers(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	args := []any{userID}
	paramIdx := 2

	cursorClause := ""
	if cursor != nil {
		cursorClause = fmt.Sprintf(" AND created_at < $%d", paramIdx)
		args = append(args, *cursor)
		paramIdx++
	}

	query := `
		SELECT follower_id, following_id, created_at
		FROM follows
		WHERE following_id = $1` + cursorClause +
		" ORDER BY created_at DESC, follower_id DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit+1)
		paramIdx++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query followers: %w", err)
	}
	defer rows.Close()

	return scanFollows(rows, limit)
}

func (r *FollowRepository) ListFollowing(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListFollowsResult, error) {
	args := []any{userID}
	paramIdx := 2

	cursorClause := ""
	if cursor != nil {
		cursorClause = fmt.Sprintf(" AND created_at < $%d", paramIdx)
		args = append(args, *cursor)
		paramIdx++
	}

	query := `
		SELECT follower_id, following_id, created_at
		FROM follows
		WHERE follower_id = $1` + cursorClause +
		" ORDER BY created_at DESC, follower_id DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit+1)
		paramIdx++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query following: %w", err)
	}
	defer rows.Close()

	return scanFollows(rows, limit)
}

func scanFollows(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}, limit int) (*domain.ListFollowsResult, error) {
	follows := make([]domain.Follow, 0)
	for rows.Next() {
		var f domain.Follow
		if err := rows.Scan(&f.FollowerID, &f.FollowingID, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan follow: %w", err)
		}
		follows = append(follows, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	var nextCursor *string
	if len(follows) > limit {
		follows = follows[:limit]
		c := follows[len(follows)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
		nextCursor = &c
	}

	return &domain.ListFollowsResult{Data: follows, NextCursor: nextCursor}, nil
}
