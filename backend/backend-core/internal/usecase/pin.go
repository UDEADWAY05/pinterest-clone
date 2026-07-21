package usecase

import (
	"context"
	"errors"
	"strings"

	"pinterest-core/internal/domain"
	repository "pinterest-core/internal/repository"
)

type PinUsecase struct {
	Repo repository.PinRepository
}

func NewPinUsecase(repo repository.PinRepository) *PinUsecase {
	return &PinUsecase{Repo: repo}
}

func (u *PinUsecase) List(ctx context.Context) ([]domain.Pin, error) {
	return u.Repo.ListPins(ctx)
}

func (u *PinUsecase) GetByID(ctx context.Context, id int64) (*domain.Pin, error) {
	return u.Repo.GetPinByID(ctx, id)
}

func (u *PinUsecase) Create(ctx context.Context, title, description, dominant_color, image_url string, width, heidth int, author_id int64, tags []string) (*domain.Pin, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	image_url = strings.TrimSpace(image_url)

	if title == "" {
		return nil, errors.New("title is required")
	}

	if image_url == "" {
		return nil, errors.New("image_url is required")
	}

	if heidth == 0 || width == 0 {
		return nil, errors.New("image size not correct")
	}

	pin := &domain.Pin{
		Title:    title,
		Description: description,
		ImageURL: image_url,
		AuthorId: author_id,
		Width:    width,
		Height:   heidth,
		Tags:     tags,
	}

	err := u.Repo.CreatePin(ctx, pin)
	if err != nil {
		return nil, err
	}

	// Сохраняем теги
	if len(tags) > 0 {
		if tagRepo, ok := u.Repo.(interface {
			SaveTags(ctx context.Context, pinID int64, tags []string) error
		}); ok {
			_ = tagRepo.SaveTags(ctx, pin.ID, tags)
		}
	}

	return pin, nil
}
