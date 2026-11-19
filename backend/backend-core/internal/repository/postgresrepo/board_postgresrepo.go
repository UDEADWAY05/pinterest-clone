package postgresrepo

import (
	"context"
	"fmt"
	"pinterest-core/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BoardRepository struct {
	Pool *pgxpool.Pool
}

func NewBoardRepository(pool *pgxpool.Pool) *BoardRepository {
	return &BoardRepository{Pool: pool}
}

func (r *BoardRepository) CreateBoard(ctx context.Context, board *domain.Board) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO boards (owner_id, title, is_private, is_default)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		board.OwnerID, board.Title, board.IsPrivate, board.IsDefault,
	).Scan(&board.ID, &board.CreatedAt)

	if err != nil {
		return fmt.Errorf("create board: %w", err)
	}
	return nil
}

func (r *BoardRepository) ListUserBoards(ctx context.Context, userID int64) ([]domain.BoardWithRecentPins, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT
    b.id,
    b.title,
    b.owner_id,
    b.is_private,
    b.created_at,
    b.is_default
    COALESCE(
        json_agg(
            json_build_object(
                'id', bp.pin_id,
                'image_url', bp.image_url,
                'saved_at', bp.saved_at
            )
        ) FILTER (WHERE bp.pin_id IS NOT NULL), '[]'
    ) AS recent_pins
		FROM boards b
		LEFT JOIN LATERAL (
    SELECT
        bp_inner.pin_id,
        bp_inner.saved_at,
        p.image_url
    FROM board_pins bp_inner
    JOIN pins p ON p.id = bp_inner.pin_id
    WHERE bp_inner.board_id = b.id
    ORDER BY bp_inner.saved_at DESC
    LIMIT 3
		) bp ON TRUE
		WHERE b.owner_id = $1
		GROUP BY b.id, b.title, b.owner_id, b.is_private, b.created_at;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user boards: %w", err)
	}
	defer rows.Close()

	var boards []domain.BoardWithRecentPins
	for rows.Next() {
		var b domain.BoardWithRecentPins
		if err := rows.Scan(&b.ID, &b.Title, &b.OwnerID, &b.IsPrivate, &b.CreatedAt, &b.IsDefault, &b.RecentPins); err != nil {
			return nil, fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, b)
	}
	return boards, nil
}

func (r *BoardRepository) GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error) {
	var b domain.Board
	err := r.Pool.QueryRow(ctx, `
		SELECT id, owner_id, title, is_private, created_at
		FROM boards WHERE id = $1`,
		boardID,
	).Scan(&b.ID, &b.OwnerID, &b.Title, &b.IsPrivate, &b.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("get board by id: %w", err)
	}
	return &b, nil
}


func (r *BoardRepository) SavePinToBoard(ctx context.Context, boardID, pinID int64) (*domain.BoardPin, error) {
	var bp domain.BoardPin
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO board_pins (board_id, pin_id)
		VALUES ($1, $2)
		RETURNING board_id, pin_id, saved_at`,
		boardID, pinID,
	).Scan(&bp.BoardID, &bp.PinID, &bp.SavedAt)

	if err != nil {
		return nil, fmt.Errorf("save pin to board: %w", err)
	}
	return &bp, nil
}

func (r *BoardRepository) RemovePinFromBoard(ctx context.Context, boardID, pinID int64) error {
	_, err := r.Pool.Exec(ctx, `
		DELETE FROM board_pins
		WHERE board_id = $1 AND pin_id = $2`,
		boardID, pinID,
	)
	if err != nil {
		return fmt.Errorf("remove pin from board: %w", err)
	}
	return nil
}

func (r *BoardRepository) GetDefaultBoard(ctx context.Context, userID int64) (*domain.Board, error) {
	var board domain.Board
	err := r.Pool.QueryRow(ctx, `
		SELECT id, owner_id, title, is_private, created_at, is_default
		FROM boards
		WHERE owner_id = $1 AND is_default = true`,
		userID,
	).Scan(&board.ID, &board.OwnerID, &board.Title, &board.IsPrivate, &board.CreatedAt, &board.IsDefault)
	if err != nil {
		return nil, fmt.Errorf("get default board: %w", err)
	}
	return &board, nil
}

func (r *BoardRepository) GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT p.id, p.title, p.description, p.image_url,
		       p.dominant_color, p.height, p.width, p.author_id, p.created_at
		FROM pins p
		JOIN board_pins bp ON p.id = bp.pin_id
		WHERE bp.board_id = $1
		ORDER BY bp.saved_at DESC`,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("get board pins: %w", err)
	}
	defer rows.Close()

	var pins []domain.Pin
	for rows.Next() {
		var pin domain.Pin
		if err := rows.Scan(&pin.ID, &pin.Title, &pin.Description, &pin.ImageURL,
			&pin.DominantColor, &pin.Height, &pin.Width, &pin.AuthorId, &pin.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan board pin: %w", err)
		}
		pins = append(pins, pin)
	}
	return pins, nil
}
