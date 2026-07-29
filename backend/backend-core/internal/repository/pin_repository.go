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
	// FindTrending возвращает самые популярные пины.
	FindTrending(ctx context.Context, limit int) ([]domain.Pin, error)
	// FindRelated возвращает пины, похожие на указанный.
	// Приоритет: по embedding. Fallback: по дате.
	FindRelated(ctx context.Context, pinID int64, limit int) ([]domain.Pin, error)
	// FindExploration возвращает пины, которые пользователь ещё не видел
	// (нет в interactions по session_id). Используется для разрушения пузыря фильтров.
	FindExploration(ctx context.Context, sessionID string, limit int) ([]domain.Pin, error)
	// LoadTagsForPins загружает теги для нескольких пинов одним запросом.
	// Нужно для MMR-переранжирования (diversity по тегам).
	LoadTagsForPins(ctx context.Context, pinIDs []int64) (map[int64][]string, error)
	// LoadTags(ctx context.Context, pinID int64) ([]string, error)
	// SaveTags(ctx context.Context, pinID int64, tags []string) error
}
