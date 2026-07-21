package repository

import (
	"context"
	"time"
)

// CacheRepository — абстракция над любым кеш-бэкендом.
//
// Использование интерфейса (а не прямого вызова Redis) даёт два плюса:
//  1. Тестируемость: в тестах можно подставить in-memory реализацию.
//  2. Заменяемость: если завтра вместо Redis понадобится, например, BigCache
//     или memcached — меняем только реализацию, не трогая бизнес-логику.
//
// Это проявление Dependency Inversion Principle (буква "D" из SOLID):
// высокоуровневые модули (usecase/repository) не зависят от низкоуровневых (конкретный драйвер Redis).
// Оба зависят от абстракции — этого интерфейса.
type CacheRepository interface {
	// Get извлекает значение по ключу.
	// Возвращает (nil, nil) если ключ не найден — это идиоматичный подход в Go.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set сохраняет значение с временем жизни ttl.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete удаляет один ключ.
	Delete(ctx context.Context, key string) error

	// DeleteByPattern удаляет все ключи, соответствующие glob-шаблону.
	// Используется для инвалидации групп ключей (например, всей ленты пользователя).
	//
	// ⚠️ ВАЖНО: KEYS * — это O(N) по всем ключам в Redis.
	// На проде лучше заменить на SCAN (курсорная итерация) или
	// хранить ключи групп в Redis SET. Но для MVP KEYS достаточно.
	DeleteByPattern(ctx context.Context, pattern string) error
}
