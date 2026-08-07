package repository

import (
	"context"
	"time"

	"pinterest-core/internal/domain"
)

type CachedCommentRepository struct {
	wrapped CommentRepository
}

func NewCachedCommentRepository(wrapped CommentRepository) *CachedCommentRepository {
	return &CachedCommentRepository{wrapped: wrapped}
}

func (r *CachedCommentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return r.wrapped.CreateComment(ctx, comment)
}

func (r *CachedCommentRepository) ListCommentsByPin(ctx context.Context, pinID int64, cursor *time.Time, limit int) (*domain.ListCommentsResult, error) {
	return r.wrapped.ListCommentsByPin(ctx, pinID, cursor, limit)
}

func (r *CachedCommentRepository) DeleteComment(ctx context.Context, commentID, authorID int64) error {
	return r.wrapped.DeleteComment(ctx, commentID, authorID)
}
