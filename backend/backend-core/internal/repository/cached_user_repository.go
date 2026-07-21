package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pinterest-core/internal/domain"
)

// CachedUserRepository — кеширующий декоратор для UserRepository.
//
// Профили пользователей — идеальные кандидаты для кеширования:
//   - Читаются часто (каждый просмотр пина тянет автора; /api/user/current).
//   - Меняются редко.
//   - Маленькие по размеру (один объект ~200 байт).
//
// Поэтому userTTL можно ставить большим (5–10 минут), что даёт
// очень высокий hit rate (процент попаданий в кеш).
type CachedUserRepository struct {
	wrapped UserRepository
	cache   CacheRepository
	ttl     time.Duration
}

func NewCachedUserRepository(
	wrapped UserRepository,
	cache CacheRepository,
	ttl time.Duration,
) *CachedUserRepository {
	return &CachedUserRepository{
		wrapped: wrapped,
		cache:   cache,
		ttl:     ttl,
	}
}

// ── Ключи ───────────────────────────────────────────────────────────────

func userByIDKey(id int64) string    { return fmt.Sprintf("user:id:%d", id) }
func userByEmailKey(email string) string { return fmt.Sprintf("user:email:%s", email) }
func userListKey() string            { return "user:list:all" }

// ═══════════════════════════════════════════════════════════════════════════════
// СТРАТЕГИЯ: Двойной индекс в кеше
//
// Для пользователей мы кешируем ДВА ключа на один объект:
//   - user:id:{ID}     — для GetUserByID
//   - user:email:{email} — для GetUserByEmail
//
// Это trade-off:
//   (+) Быстрый доступ по обоим полям (как индексы в БД, только в Redis).
//   (−) Двойное использование памяти (но для пользователей это копейки).
//   (−) При инвалидации нужно удалять оба ключа.
//
// Альтернатива: кешировать только по ID, а GetUserByEmail всегда ходить в БД.
// Но поскольку GetUserByEmail вызывается при КАЖДОМ логине, кеширование окупается.
// ═══════════════════════════════════════════════════════════════════════════════

// GetUserByID — кешированный запрос пользователя по ID.
func (r *CachedUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	key := userByIDKey(id)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user:id:%d failed: %v", id, err)
	}
	if cached != nil {
		var user domain.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
		log.Printf("cache unmarshal user:id:%d failed: %v", id, err)
	}

	user, err := r.wrapped.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, user)
	return user, nil
}

// GetUserByEmail — кешированный запрос пользователя по email.
func (r *CachedUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	key := userByEmailKey(email)

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user:email:%s failed: %v", email, err)
	}
	if cached != nil {
		var user domain.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
		log.Printf("cache unmarshal user:email:%s failed: %v", email, err)
	}

	user, err := r.wrapped.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, user)

	// Дополнительно кешируем по ID — чтобы GetUserByID тоже попал в кеш.
	// Это разумно: мы уже достали юзера из БД, давай сохраним и по ID.
	r.cacheSet(ctx, userByIDKey(user.ID), user)

	return user, nil
}

// ListUsers — кешированный список пользователей.
func (r *CachedUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	key := userListKey()

	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Printf("cache get user list failed: %v", err)
	}
	if cached != nil {
		var users []domain.User
		if err := json.Unmarshal(cached, &users); err == nil {
			return users, nil
		}
		log.Printf("cache unmarshal user list failed: %v", err)
	}

	users, err := r.wrapped.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	r.cacheSet(ctx, key, users)
	return users, nil
}

// CreateUser — создание пользователя с инвалидацией списка.
func (r *CachedUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if err := r.wrapped.CreateUser(ctx, user); err != nil {
		return err
	}

	// Инвалидируем список — новый пользователь должен появиться.
	if err := r.cache.Delete(ctx, userListKey()); err != nil {
		log.Printf("cache invalidate user list failed: %v", err)
	}

	return nil
}

func (r *CachedUserRepository) cacheSet(ctx context.Context, key string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal %s failed: %v", key, err)
		return
	}
	if err := r.cache.Set(ctx, key, data, r.ttl); err != nil {
		log.Printf("cache set %s failed: %v", key, err)
	}
}
