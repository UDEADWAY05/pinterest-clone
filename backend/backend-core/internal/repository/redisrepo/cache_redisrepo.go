package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository — конкретная реализация repository.CacheRepository на Redis.
//
// Этот тип — адаптер (Adapter pattern) между нашим интерфейсом CacheRepository
// и библиотекой go-redis. Он «переводит» обобщённые вызовы Get/Set/Delete
// в конкретные Redis-команды.
//
// Структура хранит *redis.Client, а не свою обёртку infra/redis.Redis,
// чтобы не плодить лишних зависимостей между слоями. Это упрощает тестирование:
// в юнит-тестах можно передать miniredis напрямую.
type CacheRepository struct {
	client *redis.Client
}

// NewCacheRepository создаёт новый экземпляр Redis-кеша.
func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// Get получает значение по ключу из Redis.
//
// Используем redis.Nil для обработки случая «ключ не найден»:
// возвращаем (nil, nil) — идиоматичный способ сказать «данных нет, но это не ошибка».
// Это позволяет вызывающему коду отличать «кеш пуст» от «Redis упал».
func (r *CacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Ключ не найден — не ошибка, просто пустой кеш
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get %q: %w", key, err)
	}
	return val, nil
}

// Set сохраняет значение в Redis с заданным TTL.
//
// Redis TTL — это механизм автоматического удаления ключей.
// Когда TTL истекает (например, 5 минут для ленты), Redis сам удаляет ключ.
// Это решает проблему «протухания» данных: не нужно писать отдельный
// garbage collector — Redis делает это за нас.
func (r *CacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", key, err)
	}
	return nil
}

// Delete удаляет один ключ.
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del %q: %w", key, err)
	}
	return nil
}

// DeleteByPattern удаляет все ключи по glob-шаблону.
//
// Реализация через KEYS + DEL — проста, но на крупном проде опасна
// (KEYS блокирует весь Redis на время сканирования). На нашем масштабе
// (порядка тысяч ключей) это работает быстро.
//
// Альтернатива для прода: SCAN-курсор (неблокирующий) или хранение
// групп ключей в Redis SET + SMEMBERS.
func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("redis keys %q: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del pattern %q: %w", pattern, err)
	}

	return nil
}
