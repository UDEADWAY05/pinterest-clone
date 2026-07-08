package repository

import (
	"context"
	"pinterest-core/internal/domain"
)

type PinRepository interface {
	ListPins(ctx context.Context) ([]domain.Pin, error)
	GetPinByID(ctx context.Context, id int64) (*domain.Pin, error)
	CreatePin(ctx context.Context, pin *domain.Pin) error
}
