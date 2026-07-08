package repository

import (
	"context"
	"pinterest-core/internal/domain"
)

var ErrUserNotFound = domainError("user not found")

type UserRepository interface {
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}

type domainError string

func (e domainError) Error() string { return string(e) }
