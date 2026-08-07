package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"pinterest-core/internal/domain"
	repository "pinterest-core/internal/repository"
)

type CommentUsecase struct {
	Repo repository.CommentRepository
}

func NewCommentUsecase(repo repository.CommentRepository) *CommentUsecase {
	return &CommentUsecase{Repo: repo}
}

func (u *CommentUsecase) ListByPin(ctx context.Context, pinID int64, cursor *time.Time, limit int) (*domain.ListCommentsResult, error) {
	return u.Repo.ListCommentsByPin(ctx, pinID, cursor, limit)
}

func (u *CommentUsecase) Create(ctx context.Context, pinID, authorID int64, body string) (*domain.Comment, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("body is required")
	}

	comment := &domain.Comment{
		PinID:    pinID,
		AuthorID: authorID,
		Body:     body,
	}

	if err := u.Repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (u *CommentUsecase) Delete(ctx context.Context, commentID, authorID int64) error {
	return u.Repo.DeleteComment(ctx, commentID, authorID)
}
