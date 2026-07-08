package postgresrepo

import (
	"context"
	"fmt"
	"pinterest-core/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PinRepository struct {
	Pool *pgxpool.Pool
}



func NewPinRepository(pool *pgxpool.Pool) *PinRepository {
	return &PinRepository{Pool: pool}
}


func (r *PinRepository) ListPins(ctx context.Context) ([]domain.Pin, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, title, description, image_url, dominant_color, height, width, created_at FROM pins ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()


	var pins []domain.Pin
	for rows.Next() {
		var pin domain.Pin
		if err := rows.Scan(&pin.ID, &pin.Title, &pin.Description, &pin.ImageURL, &pin.DominantColor, &pin.Height, &pin.Width, &pin.CreatedAt); err != nil {
			return nil,  fmt.Errorf("scan note: %w", err)
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func (r *PinRepository) GetPinByID(ctx context.Context, id int64) (*domain.Pin, error) {
	var pin domain.Pin

	err := r.Pool.QueryRow(ctx, `
		SELECT id, title, description, created_at, image_url, dominant_color, height, width
		FROM pins WHERE id = $1`,
		id).Scan(
		&pin.ID,
		&pin.Title,
		&pin.Description,
		&pin.CreatedAt,
		&pin.ImageURL,
		&pin.DominantColor,
		&pin.Height,
		&pin.Width,
	)

	if err != nil {
		return nil, fmt.Errorf("get pin by id: %w", err)
	}

	return &pin, nil
}

func (r *PinRepository) CreatePin(ctx context.Context, pin *domain.Pin) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO pins (title, description, image_url, dominant_color, author_id, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;`, // <-- Опечатка исправлена здесь
		pin.Title, pin.Description, pin.ImageURL, pin.DominantColor, pin.AuthorId, pin.Width, pin.Height,
	).Scan(&pin.ID, &pin.CreatedAt)

	if err != nil {
		return fmt.Errorf("create pin: %w", err)
	}
	return nil
}
