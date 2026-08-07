package postgresrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

type CommentRepository struct {
	Pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{Pool: pool}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO comments (pin_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at;`,
		comment.PinID, comment.AuthorID, comment.Body,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) ListCommentsByPin(ctx context.Context, pinID int64, cursor *time.Time, limit int) (*domain.ListCommentsResult, error) {
	args := []any{pinID}
	paramIdx := 2

	cursorClause := ""
	if cursor != nil {
		cursorClause = fmt.Sprintf(" AND created_at > $%d", paramIdx)
		args = append(args, *cursor)
		paramIdx++
	}

	query := `
		SELECT id, pin_id, author_id, body, created_at, updated_at
		FROM comments
		WHERE pin_id = $1 AND deleted_at IS NULL` + cursorClause +
		" ORDER BY created_at ASC, id ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit+1)
		paramIdx++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.PinID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	var nextCursor *string
	if len(comments) > limit {
		comments = comments[:limit]
		c := comments[len(comments)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
		nextCursor = &c
	}

	return &domain.ListCommentsResult{Data: comments, NextCursor: nextCursor}, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, commentID, authorID int64) error {
	tag, err := r.Pool.Exec(ctx, `
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL;`,
		commentID, authorID,
	)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return repository.ErrCommentNotFound
	}
	return nil
}
