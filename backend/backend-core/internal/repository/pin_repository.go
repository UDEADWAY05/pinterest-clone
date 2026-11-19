package repository

import (
	"context"
	"pinterest-core/internal/domain"
	"time"
)

type PinRepository interface {
	ListPins(ctx context.Context, cursor *time.Time, limit int, userId *int64) (*domain.ListPinsResult, error)
	GetPinByID(ctx context.Context, id int64) (*domain.Pin, error)
	GetPinByIDs(ctx context.Context, ids []int64) ([]domain.Pin, error)
	CreatePin(ctx context.Context, pin *domain.Pin) error
	DeletePin(ctx context.Context, id int64, authorID int64) error
	UpdatePin(ctx context.Context, id int64, author_id int64, pinDTO *domain.UpdatePinParams) (*domain.Pin, error)
	FindTrending(ctx context.Context, limit int) ([]domain.Pin, error)
	FindRelated(ctx context.Context, pinID int64, limit int) ([]domain.Pin, error)
	FindExploration(ctx context.Context, sessionID string, limit int) ([]domain.Pin, error)
	LoadTagsForPins(ctx context.Context, pinIDs []int64) (map[int64][]string, error)

	ListUserCreatedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error)
	ListUserSavedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error)

	Search(ctx context.Context, query string, cursor *time.Time, limit int) (*domain.ListPinsResult, error)

	EnrichPins(ctx context.Context, pins []domain.Pin, viewerID *int64) ([]domain.Pin, error)
}
