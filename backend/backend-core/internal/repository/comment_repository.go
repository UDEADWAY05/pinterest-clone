package repository

import (
	"context"
	"time"

	"pinterest-core/internal/domain"
)

var ErrCommentNotFound = domainError("comment not found")

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *domain.Comment) error
	ListCommentsByPin(ctx context.Context, pinID int64, cursor *time.Time, limit int) (*domain.ListCommentsResult, error)
	DeleteComment(ctx context.Context, commentID, authorID int64) error
}
