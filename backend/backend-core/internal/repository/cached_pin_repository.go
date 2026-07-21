package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pinterest-core/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ПАТТЕРН: Decorator (обёртка/декоратор)
//
// CachedPinRepository реализует PinRepository и оборачивает другой PinRepository
// (в нашем случае — Postgres-реализацию). Он добавляет кеширование, не меняя
// оригинальный код.
//
// Это классический пример Open/Closed Principle (буква "O" из SOLID):
//   - Код Postgres-репозитория ЗАКРЫТ для изменений (мы его не трогаем).
//   - Но ОТКРЫТ для расширения (добавляем декоратор с кешированием).
//
// ═══════════════════════════════════════════════════════════════════════════════
// СТРАТЕГИЯ КЕШИРОВАНИЯ: Cache-Aside (Lazy Loading)
//
// Это самая распространённая стратегия:
//
//  1. При чтении: проверить кеш → если есть (HIT) → вернуть из кеша.
//     → если нет (MISS) → загрузить из БД → сохранить в кеш → вернуть.
//  2. При записи: обновить БД → УДАЛИТЬ (инвалидировать) связанные ключи из кеша.
//
// Почему именно удалять, а не обновлять кеш при записи?
//   - Избегаем race condition: два параллельных обновления могут записать
//     в кеш неконсистентные данные.
//   - Инвалидация проще: следующий читатель сам загрузит свежие данные (lazy load).
//   - Не пишем в кеш данные, которые, возможно, никто не запросит.
//
// Проблема Cache Stampede (топот стада):
// Если 100 запросов придут одновременно на «холодный» ключ (после инвалидации),
// все 100 пойдут в БД. Для MVP это терпимо, на проде решается через:
//   - singleflight (golang.org/x/sync/singleflight) — выполняет только один
//     реальный запрос, остальные ждут результат.
//   - Блокировки на уровне ключа (Redis SET NX).
//
// ═══════════════════════════════════════════════════════════════════════════════

type CachedPinRepository struct {
	// wrapped — «настоящий» репозиторий, в который мы делегируем при cache miss.
	wrapped PinRepository
	// cache — абстракция над Redis (или любым другим кеш-бэкендом).
	cache CacheRepository
	// pinTTL — время жизни кеша для одного пина.
	pinTTL time.Duration
	// listTTL — время жизни кеша для списка пинов.
	listTTL time.Duration
}

// NewCachedPinRepository создаёт кеширующую обёртку.
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

// ═══════════════════════════════════════════════════════════════════════
// Формат ключей Redis
//
// Используем префиксную нотацию: "тип_сущности:идентификатор".
// Это позволяет:
//   - Легко инвалидировать группы (pin:*, pin:list:*).
//   - При отладке `redis-cli KEYS pin:*` показывает все ключи пинов.
//   - Избежать коллизий ключей между разными сущностями.

func pinKey(id int64) string  { return fmt.Sprintf("pin:%d", id) }
func pinListKey() string       { return "pin:list:all" }

// ═══════════════════════════════════════════════════════════════════════

// GetPinByID — кешированный запрос одного пина.
func (r *CachedPinRepository) GetPinByID(ctx context.Context, id int64) (*domain.Pin, error) {
	key := pinKey(id)

	// 1. Попытка достать из кеша
	cached, err := r.cache.Get(ctx, key)
	if err != nil {
		// Redis недоступен — не фатально, просто идём в БД.
		// Это называется graceful degradation: сервис продолжает работать,
		// но медленнее, без кеша.
		log.Printf("cache get pin:%d failed: %v", id, err)
	}
	if cached != nil {
		var pin domain.Pin
		if err := json.Unmarshal(cached, &pin); err == nil {
			return &pin, nil
		}
		// Если данные в кеше повреждены — логируем и идём в БД
		log.Printf("cache unmarshal pin:%d failed: %v", id, err)
	}

	// 2. Cache MISS — идём в БД
	pin, err := r.wrapped.GetPinByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в кеш для будущих запросов.
	// Ошибка при сохранении в кеш не фатальна — данные вернутся пользователю,
	// просто следующий запрос снова пойдёт в БД.
	r.cacheSet(ctx, key, pin, r.pinTTL)

	return pin, nil
}

// ListPins — кешированный список всех пинов.
//
// Это самая «горячая» точка: лента на главной странице Pinterest.
// Кеширование здесь даёт максимальный выигрыш, потому что лента:
//   1. Запрашивается чаще всего.
//   2. Дорого считается (SELECT с сортировкой по created_at).
//   3. Допускает небольшую задержку в актуальности (1 минута — ок).
func (r *CachedPinRepository) ListPins(ctx context.Context) ([]domain.Pin, error) {
	key := pinListKey()

	// 1. Попытка из кеша
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

	// 2. Загрузка из БД
	pins, err := r.wrapped.ListPins(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в кеш
	r.cacheSet(ctx, key, pins, r.listTTL)

	return pins, nil
}

// CreatePin — создание пина с инвалидацией кеша.
//
// Почему недостаточно просто записать в БД?
// После создания нового пина кеш списка ВСЕХ пинов (pin:list:all)
// содержит УСТАРЕВШИЕ данные (без нового пина). Поэтому мы его удаляем.
// Следующий читатель сам загрузит свежий список.
func (r *CachedPinRepository) CreatePin(ctx context.Context, pin *domain.Pin) error {
	// 1. Пишем в БД (основной источник правды)
	if err := r.wrapped.CreatePin(ctx, pin); err != nil {
		return err
	}

	// 2. Инвалидируем кеш списка — новый пин должен появиться в ленте.
	// Конкретный пин не кешируем при создании, дадим ему «прогреться»
	// через первый же GetPinByID.
	if err := r.cache.Delete(ctx, pinListKey()); err != nil {
		log.Printf("cache invalidate pin list failed: %v", err)
	}

	return nil
}

// cacheSet — вспомогательный метод для сохранения в кеш с обработкой ошибок.
//
// Вынесен в отдельный метод, чтобы не дублировать код во всех read-методах.
// Ошибка только логируется — мы не хотим ронять запрос из-за проблем с кешем.
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
