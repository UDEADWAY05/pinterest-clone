package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pinterest-core/internal/domain"
)

type CachedPinRepository struct {
	wrapped PinRepository
	cache CacheRepository
	pinTTL time.Duration
	listTTL time.Duration
}

func NewCachedPinRepository(
	wrapped PinRepository,
	cache CacheRepository,
	pinTTL time.Duration,
	listTTL time.Duration,
) *CachedPinRepository {
	return &CachedPinRepository{
		wrapped: wrapped,
		cache:   cache,
		pinTTL:  pinTTL,
		listTTL: listTTL,
	}
}

func pinKey(id int64) string  { return fmt.Sprintf("pin:%d", id) }
func pinListKey() string       { return "pin:list:all" }

func (r *CachedPinRepository) GetPinByID(ctx context.Context, id int64) (*domain.Pin, error) {
	key := pinKey(id)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get pin:%d failed: %v", id, err)
	}
	if cached != nil {
		var pin domain.Pin
		if err := json.Unmarshal(cached, &pin); err == nil {
			return &pin, nil
		}
		log.Printf("cache unmarshal pin:%d failed: %v", id, err)
	}

	pin, err := r.wrapped.GetPinByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, pin, r.pinTTL)

	return pin, nil
}

func (r *CachedPinRepository) ListPins(ctx context.Context) ([]domain.Pin, error) {
	key := pinListKey()

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get pin list failed: %v", err)
	}
	if cached != nil {
		var pins []domain.Pin
		if err := json.Unmarshal(cached, &pins); err == nil {
			return pins, nil
		}
		log.Printf("cache unmarshal pin list failed: %v", err)
	}

	pins, err := r.wrapped.ListPins(ctx)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, pins, r.listTTL)

	return pins, nil
}


func (r *CachedPinRepository) CreatePin(ctx context.Context, pin *domain.Pin) error {
	if err := r.wrapped.CreatePin(ctx, pin); err != nil {
		return err
	}

	if err := r.cache.Delete(ctx, pinListKey()); err != nil {
		log.Printf("cache invalidate pin list failed: %v", err)
	}

	return nil
}

func (r *CachedPinRepository) GetPinByIDs(ctx context.Context, ids []int64) ([]domain.Pin, error) {
	return r.wrapped.GetPinByIDs(ctx, ids)
}

func (r *CachedPinRepository) FindTrending(ctx context.Context, limit int) ([]domain.Pin, error) {
	return r.wrapped.FindTrending(ctx, limit)
}

func (r *CachedPinRepository) FindRelated(ctx context.Context, pinID int64, limit int) ([]domain.Pin, error) {
	return r.wrapped.FindRelated(ctx, pinID, limit)
}

func (r *CachedPinRepository) FindExploration(ctx context.Context, sessionID string, limit int) ([]domain.Pin, error) {
	return r.wrapped.FindExploration(ctx, sessionID, limit)
}

func (r *CachedPinRepository) LoadTagsForPins(ctx context.Context, pinIDs []int64) (map[int64][]string, error) {
	return r.wrapped.LoadTagsForPins(ctx, pinIDs)
}

func (r *CachedPinRepository) cacheSet(ctx context.Context, key string, value any, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal %s failed: %v", key, err)
		return
	}
	if err := r.cache.Set(ctx, key, data, ttl); err != nil {
		log.Printf("cache set %s failed: %v", key, err)
	}
}
