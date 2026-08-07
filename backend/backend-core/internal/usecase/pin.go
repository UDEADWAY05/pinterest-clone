package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"pinterest-core/internal/domain"
	repository "pinterest-core/internal/repository"
)

type PinUsecase struct {
	Repo repository.PinRepository
}

func NewPinUsecase(repo repository.PinRepository) *PinUsecase {
	return &PinUsecase{Repo: repo}
}

func (u *PinUsecase) List(ctx context.Context, cursor *time.Time, limit int, userId *int64) (*domain.ListPinsResult, error) {
	return u.Repo.ListPins(ctx, cursor, limit, userId)
}

func (u *PinUsecase) GetByID(ctx context.Context, id int64) (*domain.Pin, error) {
	return u.Repo.GetPinByID(ctx, id)
}

func (u *PinUsecase) Create(ctx context.Context, title, description, dominant_color, image_url string, width, height int, author_id int64, tags []string) (*domain.Pin, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	image_url = strings.TrimSpace(image_url)

	if title == "" {
		return nil, errors.New("title is required")
	}

	if image_url == "" {
		return nil, errors.New("image_url is required")
	}

	if height == 0 || width == 0 {
		return nil, errors.New("image size not correct")
	}

	pin := &domain.Pin{
		Title:         title,
		Description:   description,
		ImageURL:      image_url,
		DominantColor: dominant_color,
		AuthorId:      author_id,
		Width:         width,
		Height:        height,
		Tags:          tags,
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

func (u *PinUsecase) Update(ctx context.Context, id int64, authorID int64, params *domain.UpdatePinParams) (*domain.Pin, error) {
	return u.Repo.UpdatePin(ctx, id, authorID, params)
}

func (u *PinUsecase) Delete(ctx context.Context, id int64, authorID int64) error {
	return u.Repo.DeletePin(ctx, id, authorID)
}

func (u *PinUsecase) ListUserCreatedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error) {
	return u.Repo.ListUserCreatedPins(ctx, userID, cursor, limit)
}

func (u *PinUsecase) ListUserSavedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error) {
	return u.Repo.ListUserSavedPins(ctx, userID, cursor, limit)
}

func (u *PinUsecase) Search(ctx context.Context, query string, cursor *time.Time, limit int) (*domain.ListPinsResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is required")
	}
	return u.Repo.Search(ctx, query, cursor, limit)
}

func (u *PinUsecase) EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error) {
	return u.Repo.EnrichPins(ctx, pins, viewerID)
}
