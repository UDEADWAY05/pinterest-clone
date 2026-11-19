package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pinterest-core/internal/domain"
)

type CachedBoardRepository struct {
	wrapped  BoardRepository
	cache    CacheRepository
	listTTL  time.Duration
	boardTTL time.Duration
}

func NewCachedBoardRepository(
	wrapped BoardRepository,
	cache CacheRepository,
	listTTL time.Duration,
	boardTTL time.Duration,
) *CachedBoardRepository {
	return &CachedBoardRepository{
		wrapped:  wrapped,
		cache:    cache,
		listTTL:  listTTL,
		boardTTL: boardTTL,
	}
}

func userBoardsKey(userID int64) string  { return fmt.Sprintf("boards:user:%d", userID) }
func boardPinsKey(boardID int64) string  { return fmt.Sprintf("boards:%d:pins", boardID) }
func boardKey(boardID int64) string      { return fmt.Sprintf("boards:%d", boardID) }
func defaultBoardKey(userID int64) string { return fmt.Sprintf("boards:user:%d:default", userID) }


func (r *CachedBoardRepository) CreateBoard(ctx context.Context, board *domain.Board) error {
	if err := r.wrapped.CreateBoard(ctx, board); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, userBoardsKey(board.OwnerID)); err != nil {
		log.Printf("cache invalidate user boards %d: %v", board.OwnerID, err)
	}
	return nil
}

func (r *CachedBoardRepository) GetDefaultBoard(ctx context.Context, userID int64) (*domain.Board, error) {
	key := defaultBoardKey(userID)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get default board %d: %v", userID, err)
	}
	if cached != nil {
		var board *domain.Board
		if err := json.Unmarshal(cached, &board); err != nil {
			return nil, err
		}
		return board, nil
	}

	board, err := r.wrapped.GetDefaultBoard(ctx, userID)
	if err != nil {
		return nil, err
	}

	r.cacheSetJSON(ctx, key, board, r.boardTTL)
	return board, nil
}

//NOTE: Нужно подумать как правильно это кешировать и есть ли вообще в этом смысл сейчас при добавлении пина на доску он сразу не появляется в recently pinned доски
func (r *CachedBoardRepository) ListUserBoards(ctx context.Context, userID int64) ([]domain.BoardWithRecentPins, error) {
	key := userBoardsKey(userID)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get boards:user:%d: %v", userID, err)
	}
	if cached != nil {
		var boards []domain.BoardWithRecentPins
		if err := json.Unmarshal(cached, &boards); err == nil {
			return boards, nil
		}
		log.Printf("cache unmarshal boards:user:%d: %v", userID, err)
	}

	boards, err := r.wrapped.ListUserBoards(ctx, userID)
	if err != nil {
		return nil, err
	}

	// r.cacheSetJSON(ctx, key, boards, r.listTTL)
	return boards, nil
}

func (r *CachedBoardRepository) GetBoardByID(ctx context.Context, boardID int64) (*domain.Board, error) {
	key := boardKey(boardID)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get board %d: %v", boardID, err)
	}
	if cached != nil {
		var b domain.Board
		if err := json.Unmarshal(cached, &b); err == nil {
			return &b, nil
		}
		log.Printf("cache unmarshal board %d: %v", boardID, err)
	}

	b, err := r.wrapped.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	r.cacheSetJSON(ctx, key, b, r.boardTTL)
	return b, nil
}

func (r *CachedBoardRepository) SavePinToBoard(ctx context.Context, boardID, pinID int64) (*domain.BoardPin, error) {
	bp, err := r.wrapped.SavePinToBoard(ctx, boardID, pinID)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Delete(ctx, boardPinsKey(boardID)); err != nil {
		log.Printf("cache invalidate board pins %d: %v", boardID, err)
	}
	return bp, nil
}

func (r *CachedBoardRepository) RemovePinFromBoard(ctx context.Context, boardID, pinID int64) error {
	if err := r.wrapped.RemovePinFromBoard(ctx, boardID, pinID); err != nil {
		return err
	}
	if err := r.cache.Delete(ctx, boardPinsKey(boardID)); err != nil {
		log.Printf("cache invalidate board pins %d: %v", boardID, err)
	}
	return nil
}

func (r *CachedBoardRepository) GetBoardPins(ctx context.Context, boardID int64) ([]domain.Pin, error) {
	key := boardPinsKey(boardID)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get board pins %d: %v", boardID, err)
	}
	if cached != nil {
		var pins []domain.Pin
		if err := json.Unmarshal(cached, &pins); err == nil {
			return pins, nil
		}
		log.Printf("cache unmarshal board pins %d: %v", boardID, err)
	}

	pins, err := r.wrapped.GetBoardPins(ctx, boardID)
	if err != nil {
		return nil, err
	}

	r.cacheSetJSON(ctx, key, pins, r.listTTL)
	return pins, nil
}

func (r *CachedBoardRepository) cacheSetJSON(ctx context.Context, key string, value any, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal %s: %v", key, err)
		return
	}
	if err := r.cache.Set(ctx, key, data, ttl); err != nil {
		log.Printf("cache set %s: %v", key, err)
	}
}
