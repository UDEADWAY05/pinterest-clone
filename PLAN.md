# PLAN: Pinterest Clone (Next.js + Go + ML)

## 1. Цель проекта

Собрать MVP-клон Pinterest с сильной базой для роста:

- Лента и просмотр пинов.
- Сохранение пинов в доски.
- Авторизация и профиль.
- Загрузка собственных пинов.
- Подготовленная архитектура под систему рекомендаций.

Проект строится в 2 слоя:

1. Сначала надежный продуктовый core (frontend + backend).
2. Затем рекомендации как отдельный ML-контур без блокировки MVP.

---

## 2. Принципы архитектуры backend

### 2.1 Стратегия

На старте: модульный монолит на Go.

- Один сервис быстрее собирать и деплоить.
- Границы модулей сразу проектируются так, чтобы потом вынести их в отдельные сервисы без ломки API.

### 2.2 Архитектурный стиль

Рекомендуемый стиль: Clean/Hexagonal (упрощенный).

- transport: HTTP handlers, middleware, DTO.
- usecase/service: бизнес-логика.
- repository: работа с БД/кэшем.
- domain: сущности и правила.

### 2.3 Нефункциональные требования

- Явная валидация входных данных.
- Единый формат ошибок API.
- Наблюдаемость: логирование, базовые метрики, трассировка запросов.
- Безопасность: JWT/сессии, CORS whitelist, rate limiting.
- Идемпотентность для критичных endpoint-ов (например, сохранение пина).

---

## 3. Целевая структура backend

Текущий backend находится в папке backend/backend-core/main.go как прототип.
Нужно перейти к структуре:

backend/
backend-core/
cmd/
api/
main.go
internal/
config/
transport/
http/
router.go
middleware/
handlers/
pin_handler.go
board_handler.go
user_handler.go
auth_handler.go
upload_handler.go
interaction_handler.go
recommendation_handler.go
domain/
pin/
board/
user/
interaction/
recommendation/
usecase/
repository/
postgres/
redis/
infra/
storage/
queue/
mlclient/
migrations/
tests/
ml-worker/
cmd/
internal/
training/
serving/
featurestore/
ranking/

Примечание: сначала можно добавить папки по мере реализации, без полного scaffolding за один коммит.

---

## 4. Модель данных (MVP)

### 4.1 Основные сущности

- users: профиль пользователя.
- pins: пины и их метаданные.
- boards: доски пользователя.
- board_pins: связь many-to-many между доской и пином.
- pin_assets: ссылки на оригинал/превью.
- interactions: события поведения пользователя.

### 4.2 Минимальные таблицы

- users(id, email, password_hash, display_name, avatar_url, created_at)
- pins(id, author_id, title, description, image_url, width, height, dominant_color, created_at)
- boards(id, owner_id, title, is_private, created_at)
- board_pins(board_id, pin_id, saved_at)
- pin_tags(pin_id, tag)
- interactions(id, user_id nullable, session_id, pin_id, event_type, ts, meta_json)

### 4.3 Event taxonomy

event_type:

- impression
- click
- save
- unsave
- search
- open_pin
- dwell

Это важно зафиксировать в backend-контракте до начала ML-модуля.

---

## 5. API-контракты (MVP + рекомендационный контур)

### 5.1 Pins

- GET /api/pins?cursor=&limit=&sort=
- GET /api/pins/{id}
- POST /api/pins (auth)
- GET /api/users/{id}/pins

### 5.2 Boards/Saves

- GET /api/boards/me (auth)
- POST /api/boards (auth)
- POST /api/boards/{boardId}/pins/{pinId} (auth)
- DELETE /api/boards/{boardId}/pins/{pinId} (auth)

### 5.3 Auth/User

- POST /api/auth/register
- POST /api/auth/login
- POST /api/auth/logout
- GET /api/user/current (auth)

### 5.4 Upload

- POST /api/uploads/presign (auth)
- POST /api/pins (auth, после загрузки image)

### 5.5 Interactions (для рекомендаций)

- POST /api/interactions
  body: { event_type, pin_id, session_id, ts, meta }

### 5.6 Recommendations

- GET /api/recommendations?user_id=&limit=&cursor=
- GET /api/pins/{id}/related?limit=

Для деградации: при недоступности ML-сервиса возвращать fallback-ленту без ошибки UI.

---

## 6. Поток данных рекомендаций

### 6.1 Online path

1. Frontend открывает home/feed.
2. Backend запрашивает кандидатов в recommendation module (локально или через ml-worker API).
3. Ranking возвращает список pin_id + score.
4. Backend гидратирует pin_id до pin DTO.
5. Frontend получает рекомендации как отдельный блок в ленте.

### 6.2 Offline path

1. Interactions пишутся в хранилище событий.
2. ETL собирает датасет (user, pin, context, label).
3. ML-worker тренирует модель ранжирования.
4. Артефакт модели публикуется в serving-слой.
5. Online ranking переключается на новую версию по feature flag.

### 6.3 Первая версия ранжирования

Гибридный подход:

- Candidate generation: похожесть по тегам/категориям/эмбеддингам изображения.
- Ranking: логистическая регрессия или lightweight GBDT по сигналам (CTR, save-rate, recency, affinity).

---

## 7. Инфраструктура и стек

### 7.1 База и кэш

- PostgreSQL: основная OLTP БД.
- Redis: кэш, rate limit, короткие сессии/фичи.

### 7.2 Хранилище изображений

- S3-совместимое хранилище (на dev можно MinIO).
- Превью/ресайзы можно добавить отдельным worker-ом после MVP.

### 7.3 Очередь/события (этапно)

- Этап 1: запись interactions напрямую в PostgreSQL.
- Этап 2: Redis Streams или Kafka для масштабирования.

### 7.4 Наблюдаемость

- Structured logs (JSON).
- Метрики: RPS, latency p95, error_rate, recommendation_hit_rate.
- Tracing: опционально OpenTelemetry после MVP API.

---

## 8. Этапы реализации

### Этап A (1-2 недели): Backend foundation

- Вынести роутинг и handler-ы из main.go.
- Добавить config + env.
- Ввести единый response/error формат.
- Подключить PostgreSQL, миграции, базовые repository.

Результат: стабильный API-каркас, готовый к расширению.

### Этап B (1-2 недели): MVP API

- Pins CRUD (минимально: list/detail/create).
- Auth (register/login/current).
- Boards + save/unsave.
- Upload presign flow.

Результат: frontend может собрать полноценные пользовательские сценарии.

### Этап C (3-5 дней): Event instrumentation

- Endpoint /api/interactions.
- Серверная валидация и батч-оптимизация записи.
- Простейший аналитический срез по событиям.

Результат: появляются данные для рекомендаций.

### Этап D (1-2 недели): Recommendation v1

- В ml-worker реализовать candidate + ranking.
- Интегрировать backend endpoint /api/recommendations.
- Добавить fallback стратегию.

Результат: первая персонализированная лента.

### Этап E (непрерывно): Quality loop

- A/B или shadow режим ранжирования.
- Метрики CTR/save-rate/retention.
- Регулярное обновление фич и модели.

---

## 9. Definition of Done по этапам

Готово, если:

- API покрыт smoke/integration тестами.
- Лента и ключевые сценарии работают без моков.
- Ошибки стандартизированы и наблюдаемы.
- Recommendations не ломают home при деградации.
- Документация API и контракты обновлены.

---

## 10. Риски и как снизить

- Риск: рано уйти в сложный ML и задержать MVP.
  Митигация: сначала event pipeline + simple ranking.

- Риск: дрейф DTO между frontend и backend.
  Митигация: единые схемы и контрактные тесты.

- Риск: проблемы с производительностью изображений.
  Митигация: presign + CDN + lazy generation previews.

- Риск: переусложнение микросервисами.
  Митигация: модульный монолит до достижения реальной нагрузки.

---

## 11. Как использовать этот PLAN как инструкцию для Copilot в VS Code

### Вариант 1 (быстрый, вручную в чат)

Перед задачей пишите в Copilot Chat:
"Работай строго по PLAN.md из корня репозитория. Сначала покажи, какой пункт плана реализуешь, затем внеси изменения и проверь их."

### Вариант 2 (рекомендуется, постоянная инструкция)

Используйте файл .github/copilot-instructions.md (уже создан в этом репозитории).
Туда вынесено правило всегда сверяться с PLAN.md и следовать фазам.

### Вариант 3 (шаблон промпта /plan)

Создайте reusable prompt в VS Code User Prompts и используйте его как префикс для задач:

- Ссылка на план.
- Выбор фазы.
- Критерии проверки.

Рекомендуемый шаблон:

1. Прочитай PLAN.md.
2. Выбери следующий невыполненный пункт из текущей фазы.
3. Внеси минимальные изменения в код.
4. Проверь сборку/линт/тесты для измененных модулей.
5. Кратко отчитайся: что сделано, что осталось, какие риски.

---

## 12. Следующий практический шаг

Сделать рефакторинг backend/backend-core/main.go на router + handlers и подготовить первый пакет миграций PostgreSQL. Это даст наибольший выигрыш в управляемости кода уже в ближайшем спринте.
