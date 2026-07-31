package postgresrepo

import (
	"context"
	"errors"
	"fmt"
	"pinterest-core/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PinRepository struct {
	Pool *pgxpool.Pool
}

func NewPinRepository(pool *pgxpool.Pool) *PinRepository {
	return &PinRepository{Pool: pool}
}



func (r *PinRepository) ListPins(ctx context.Context, cursor *time.Time, limit int, userId *int64) (*domain.ListPinsResult, error) {
	whereClause := " WHERE deleted_at IS NULL"
	paramIdx := 1
	args := []any{}

	if userId != nil {
		whereClause += fmt.Sprintf(" AND author_id = $%d", paramIdx)
		args = append(args, *userId)
		paramIdx++
	}

	if cursor != nil {
		whereClause += fmt.Sprintf(" AND created_at < $%d", paramIdx)
		args = append(args, cursor)
		paramIdx++
	}

	query := `
		SELECT
			id, title, description, image_url, dominant_color, height, width, author_id, created_at
		FROM pins` + whereClause + " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit + 1)
		paramIdx++
	}
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query pins: %w", err)
	}
	defer rows.Close()


	pins := make([]domain.Pin, 0)
	for rows.Next() {
		var pin domain.Pin
		if err := rows.Scan(&pin.ID, &pin.Title, &pin.Description, &pin.ImageURL, &pin.DominantColor, &pin.Height, &pin.Width, &pin.AuthorId, &pin.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		pins = append(pins, pin)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	var nextCursor *string
	if len(pins) > limit {
		pins = pins[:limit]
		c := pins[len(pins)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
		nextCursor = &c
	} else {
		nextCursor = nil
	}


	return &domain.ListPinsResult{
		Data:       pins,
		NextCursor: nextCursor,
	}, nil
}

func (r *PinRepository) GetPinByID(ctx context.Context, id int64) (*domain.Pin, error) {
	var pin domain.Pin

	err := r.Pool.QueryRow(ctx, `
		SELECT id, title, description, created_at, image_url, dominant_color, height, width, author_id
		FROM pins WHERE id = $1 AND deleted_at IS NULL`,
		id).Scan(
		&pin.ID,
		&pin.Title,
		&pin.Description,
		&pin.CreatedAt,
		&pin.ImageURL,
		&pin.DominantColor,
		&pin.Height,
		&pin.Width,
		&pin.AuthorId,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get pin by id: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get pin by id: %w", err)
	}

	return &pin, nil
}

func (r *PinRepository) GetPinByIDs(ctx context.Context, ids []int64) ([]domain.Pin, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.Pool.Query(ctx, `
		SELECT id, title, description, image_url,
		       dominant_color, height, width, author_id, created_at
		FROM pins
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY created_at DESC`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("get pins by ids: %w", err)
	}
	defer rows.Close()

	return scanPins(rows)
}

func (r *PinRepository) CreatePin(ctx context.Context, pin *domain.Pin) error {
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO pins (title, description, image_url, dominant_color, author_id, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;`,
		pin.Title, pin.Description, pin.ImageURL, pin.DominantColor, pin.AuthorId, pin.Width, pin.Height,
	).Scan(&pin.ID, &pin.CreatedAt)

	if err != nil {
		return fmt.Errorf("create pin: %w", err)
	}
	return nil
}

func (r *PinRepository) DeletePin(ctx context.Context, id int64, authorID int64) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE pins
		SET deleted_at = NOW()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL;`,
		id, authorID,
	)

	if err != nil {
		return fmt.Errorf("delete pin: %w", err)
	}
	return nil
}

func (r *PinRepository) UpdatePin(ctx context.Context, id int64, author_id int64, pin *domain.UpdatePinParams) (*domain.Pin, error) {
	setValues := make([]string, 0)
	args := make([]any, 0)

	argID := 1

	if pin.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title = $%d", argID))
		args = append(args, *pin.Title)
		argID++
	}

	if pin.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description = $%d", argID))
		args = append(args, *pin.Description)
		argID++
	}

	if pin.ImageURL != nil {
		setValues = append(setValues, fmt.Sprintf("image_url = $%d", argID))
		args = append(args, *pin.ImageURL)
		argID++
	}

	if len(setValues) == 0 {
		return nil, fmt.Errorf("update pin: no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE pins
		SET %s
		WHERE id = $%d AND author_id = $%d AND deleted_at IS NULL
		RETURNING id, title, description, image_url, dominant_color, height, width, author_id, created_at`,
		strings.Join(setValues, ", "),
		argID,
		argID + 1,
	)

	args = append(args, id)
	args = append(args, author_id)

	var updatedPin domain.Pin

	err := r.Pool.QueryRow(ctx, query, args...).Scan(&updatedPin.ID, &updatedPin.Title, &updatedPin.Description, &updatedPin.ImageURL,
		&updatedPin.DominantColor, &updatedPin.Height, &updatedPin.Width, &updatedPin.AuthorId, &updatedPin.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update pin: %w", err)
	}
	return &updatedPin, nil
}

func (r *PinRepository) FindTrending(ctx context.Context, limit int) ([]domain.Pin, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT p.id, p.title, p.description, p.image_url,
		       p.dominant_color, p.height, p.width, p.author_id, p.created_at
		FROM pins p
		LEFT JOIN interactions i ON i.pin_id = p.id
			AND i.event_type IN ('save', 'click', 'open_pin')
		GROUP BY p.id
		ORDER BY COUNT(i.id) DESC, p.created_at DESC
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query trending: %w", err)
	}
	defer rows.Close()

	return scanPins(rows)
}

func (r *PinRepository) FindRelated(ctx context.Context, pinID int64, limit int) ([]domain.Pin, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT p.id, p.title, p.description, p.image_url,
		       p.dominant_color, p.height, p.width, p.author_id, p.created_at
		FROM pins p
		WHERE p.embedding IS NOT NULL
		  AND p.id != $1
		ORDER BY p.embedding <=> (SELECT embedding FROM pins WHERE id = $1)
		LIMIT $2`,
		pinID, limit,
	)

	if err == nil {
		defer rows.Close()
		return scanPins(rows)
	}

	return r.findRecentExcept(ctx, pinID, limit)
}

func (r *PinRepository) findRecentExcept(ctx context.Context, exceptID int64, limit int) ([]domain.Pin, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, title, description, image_url,
		       dominant_color, height, width, author_id, created_at
		FROM pins
		WHERE id != $1
		ORDER BY created_at DESC
		LIMIT $2`,
		exceptID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query recent pins (fallback): %w", err)
	}
	defer rows.Close()

	return scanPins(rows)
}

func (r *PinRepository) SaveTags(ctx context.Context, pinID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	copyFrom := make([][]interface{}, len(tags))
	for i, tag := range tags {
		copyFrom[i] = []interface{}{pinID, tag}
	}

	_, err := r.Pool.CopyFrom(
		ctx,
		[]string{"pin_tags"},
		[]string{"pin_id", "tag"},
		pgx.CopyFromRows(copyFrom),
	)
	if err != nil {
		return fmt.Errorf("save tags for pin %d: %w", pinID, err)
	}
	return nil
}

func (r *PinRepository) LoadTags(ctx context.Context, pinID int64) ([]string, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT tag FROM pin_tags WHERE pin_id = $1 ORDER BY tag`, pinID)
	if err != nil {
		return nil, fmt.Errorf("load tags for pin %d: %w", pinID, err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *PinRepository) LoadTagsForPins(ctx context.Context, pinIDs []int64) (map[int64][]string, error) {
	if len(pinIDs) == 0 {
		return nil, nil
	}

	rows, err := r.Pool.Query(ctx, `
		SELECT pin_id, tag FROM pin_tags
		WHERE pin_id = ANY($1)
		ORDER BY pin_id, tag`, pinIDs)
	if err != nil {
		return nil, fmt.Errorf("load tags for pins: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]string)
	for rows.Next() {
		var pinID int64
		var tag string
		if err := rows.Scan(&pinID, &tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		result[pinID] = append(result[pinID], tag)
	}
	return result, nil
}

func (r *PinRepository) FindExploration(ctx context.Context, sessionID string, limit int) ([]domain.Pin, error) {
	query := `
		SELECT id, title, description, image_url,
		       dominant_color, height, width, author_id, created_at
		FROM pins`

	if sessionID != "" {
		query += `
		WHERE id NOT IN (
			SELECT pin_id FROM interactions WHERE session_id = $1
		)`
	}

	query += `
	ORDER BY RANDOM()
	LIMIT $2`

	var rows pgx.Rows
	var err error
	if sessionID != "" {
		rows, err = r.Pool.Query(ctx, query, sessionID, limit)
	} else {
		rows, err = r.Pool.Query(ctx, query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("query exploration: %w", err)
	}
	defer rows.Close()

	return scanPins(rows)
}

func scanPins(rows pgx.Rows) ([]domain.Pin, error) {
	var pins []domain.Pin
	for rows.Next() {
		var pin domain.Pin
		if err := rows.Scan(&pin.ID, &pin.Title, &pin.Description, &pin.ImageURL,
			&pin.DominantColor, &pin.Height, &pin.Width, &pin.AuthorId, &pin.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan pin: %w", err)
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

func (r *PinRepository) ListUserCreatedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error) {
	whereClause := " WHERE author_id = $1 AND deleted_at IS NULL"
	args := []any{userID}
	paramIdx := 2

	if cursor != nil {
		whereClause += fmt.Sprintf(" AND created_at < $%d", paramIdx)
		args = append(args, *cursor)
		paramIdx++
	}

	query := `
		SELECT
			id, title, description, image_url, dominant_color, height, width, author_id, created_at
		FROM pins` + whereClause + " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit+1)
		paramIdx++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query user created pins: %w", err)
	}
	defer rows.Close()

	pins, err := scanPins(rows)
	if err != nil {
		return nil, err
	}

	var nextCursor *string
	if len(pins) > limit {
		pins = pins[:limit]
		c := pins[len(pins)-1].CreatedAt.Format(time.RFC3339Nano)
		nextCursor = &c
	}

	return &domain.ListPinsResult{Data: pins, NextCursor: nextCursor}, nil
}

func (r *PinRepository) ListUserSavedPins(ctx context.Context, userID int64, cursor *time.Time, limit int) (*domain.ListPinsResult, error) {
	args := []any{userID}
	paramIdx := 2

	cursorClause := ""
	if cursor != nil {
		cursorClause = fmt.Sprintf(" AND sp.saved_at < $%d", paramIdx)
		args = append(args, *cursor)
		paramIdx++
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.description, p.image_url, p.dominant_color, p.height, p.width, p.author_id, p.created_at, sp.saved_at
		FROM pins p
		JOIN (
			SELECT DISTINCT ON (bp.pin_id) bp.pin_id, bp.saved_at
			FROM board_pins bp
			JOIN boards b ON b.id = bp.board_id
			WHERE b.owner_id = $1
			ORDER BY bp.pin_id, bp.saved_at DESC
		) sp ON sp.pin_id = p.id
		WHERE p.deleted_at IS NULL%s
		ORDER BY sp.saved_at DESC`, cursorClause)

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, limit+1)
		paramIdx++
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query user saved pins: %w", err)
	}
	defer rows.Close()

	pins := make([]domain.Pin, 0)
	for rows.Next() {
		var pin domain.Pin
		var savedAt time.Time
		if err := rows.Scan(&pin.ID, &pin.Title, &pin.Description, &pin.ImageURL,
			&pin.DominantColor, &pin.Height, &pin.Width, &pin.AuthorId, &pin.CreatedAt, &savedAt); err != nil {
			return nil, fmt.Errorf("scan saved pin: %w", err)
		}
		pin.SavedAt = &savedAt
		pins = append(pins, pin)
	}

	var nextCursor *string
	if len(pins) > limit {
		pins = pins[:limit]
		c := pins[len(pins)-1].SavedAt.Format(time.RFC3339Nano)
		nextCursor = &c
	}

	return &domain.ListPinsResult{Data: pins, NextCursor: nextCursor}, nil
}
