package postgresrepo

import (
	"context"
	"fmt"
	"pinterest-core/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InteractionRepository struct {
	Pool *pgxpool.Pool
}

func NewInteractionRepository(pool *pgxpool.Pool) *InteractionRepository {
	return &InteractionRepository{Pool: pool}
}

// BatchSave сохраняет пачку событий одним COPY-запросом.

func (r *InteractionRepository) BatchSave(ctx context.Context, interactions []domain.Interaction) error {
	if len(interactions) == 0 {
		return nil
	}

	copyFrom := make([][]interface{}, len(interactions))
	for i, item := range interactions {
		copyFrom[i] = []interface{}{
			item.UserID,
			item.SessionID,
			item.PinID,
			item.EventType,
			item.TS,
			item.MetaJSON,
		}
	}

	_, err := r.Pool.CopyFrom(
		ctx,
		[]string{"interactions"},
		[]string{"user_id", "session_id", "pin_id", "event_type", "ts", "meta_json"},
		pgx.CopyFromRows(copyFrom),
	)
	if err != nil {
		return fmt.Errorf("copy interactions: %w", err)
	}

	return nil
}
