package usecase

import (
	"context"
	"errors"
	"strings"

	"pinterest-core/internal/domain"
	repository "pinterest-core/internal/repository"
)

type UserUsecase struct {
	Repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{Repo: repo}
}

func (u *UserUsecase) List(ctx context.Context) ([]domain.User, error) {
	return u.Repo.ListUsers(ctx)
}

func (u *UserUsecase) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return u.Repo.GetUserByID(ctx, id)
}

func (u *UserUsecase) Create(ctx context.Context, email, display_name, avatar_url string) (*domain.User, error) {
		email = strings.TrimSpace(email)
		display_name = strings.TrimSpace(display_name)
		avatar_url = strings.TrimSpace(avatar_url)

		if email == "" {
			return nil, errors.New("title is required")
		}

		if display_name == "" {
			return nil, errors.New("display_name is required")
		}

		if avatar_url == "" {
			return nil, errors.New("avatar_url is required")
		}
		
		user := &domain.User{
			Email: email,
			DisplayName: display_name,
			AvatarURL: avatar_url,
		}

		err := u.Repo.CreateUser(ctx, user)

		if err != nil {
			return nil, err
		}

		return user, nil
}
