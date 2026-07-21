package repository

import (
	"context"
	"pinterest-core/internal/domain"
)

type InteractionRepository interface {
	BatchSave(ctx context.Context, interactions []domain.Interaction) error
}
