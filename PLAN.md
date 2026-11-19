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

## 11. Следующий практический шаг

Сделать рефакторинг backend/backend-core/main.go на router + handlers и подготовить первый пакет миграций PostgreSQL. Это даст наибольший выигрыш в управляемости кода уже в ближайшем спринте.

---

## 12. Backend-доработки, необходимые фронтенду

Текущий backend покрывает базовый CRUD и auth, но для полноценного фронтенда не хватает нескольких endpoint-ов.

### 12.1 Пагинация пинов

`ListPins` сейчас отдаёт все пины без пагинации. Нужно:

- Добавить cursor-based пагинацию в `usecase.PinUsecase` и `repository.PinRepository`
- Query params: `?cursor=<ISO8601 created_at>&limit=<int>` (default 50)
- Изменить `GET /api/pins` на пагинированный ответ

### 12.2 User profile endpoints

Нужны два публичных endpoint-а:

- `GET /api/users/{id}` — профиль пользователя
  - enriched: `pins_count`, `followers_count`, `following_count`, `is_following` (если session cookie есть)
- `GET /api/users/{id}/pins?cursor=&limit=` — пины пользователя

### 12.3 Исправление багов роутинга

- router.go — добавить `DELETE /api/boards/{boardId}/pins/{pinId}` (handler существует, роут не зарегистрирован)

### 12.4 Comments (MVP, новая таблица)

- Новая таблица `comments(id BIGSERIAL, pin_id BIGINT, author_id BIGINT, body TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, deleted_at TIMESTAMPTZ)`
- FK: pin_id → pins(id), author_id → users(id)
- Endpoints:
  - `GET /api/pins/{id}/comments?cursor=&limit=` (sorted by created_at ASC)
  - `POST /api/pins/{id}/comments` (auth) — body: `{ body }`
  - `DELETE /api/comments/{id}` (auth, only owner)
- Доменная модель: `domain.Comment`
- Usecase: `CommentUsecase`
- Repository: `postgresrepo.CommentRepository` + `CachedCommentRepository`

### 12.5 Follow system (MVP, новая таблица)

- Новая таблица: `follows(follower_id BIGINT, following_id BIGINT, created_at TIMESTAMPTZ, UNIQUE(follower_id, following_id))`
- FK: follower_id → users(id), following_id → users(id)
- Endpoints:
  - `POST /api/users/{id}/follow` (auth)
  - `DELETE /api/users/{id}/follow` (auth)
  - `GET /api/users/{id}/followers?cursor=&limit=`
  - `GET /api/users/{id}/following?cursor=&limit=`
- Доменная модель: `domain.Follow`
- Usecase: `FollowUsecase`
- Repository: `postgresrepo.FollowRepository` + `CachedFollowRepository`

### 12.6 Search endpoint

- `GET /api/search?q=<string>&cursor=&limit=`
- Full-text search по `pins.title` + `pin_tags.tag`
- Fallback: `WHERE p.title ILIKE '%q%' OR pt.tag ILIKE '%q%'`
- Новый handler + usecase (можно в PinHandler или SearchHandler)

### 12.7 Обогатить Pin DTO

Текущий Pin DTO на фронтенде:

```ts
interface Pin {
  id;
  title;
  image_url;
  description;
  width;
  height;
  dominant_color;
}
```

Нужно добавить поля для UI:

```ts
interface Pin {
  id: number;
  title: string;
  image_url: string;
  description: string;
  width: number;
  height: number;
  dominant_color: string;
  author: Pick<User, "id" | "display_name" | "avatar_url">;
  tags: string[];
  saves_count: number;
  is_saved: boolean; // сохранён ли текущим пользователем
  likes_count: number;
}
```

- Backend: JOIN с `board_pins` (если session cookie), JOIN с `interactions` для count(*), JOIN с `users` для автора
- `CachedPinRepository` — инвалидация `is_saved` у каждого пользователя разная (не кешировать глобально)

### 12.8 Migration summary

```sql
-- 000005_comments
CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    pin_id BIGINT NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_comments_pin_id ON comments(pin_id, created_at);

-- 000006_follows
CREATE TABLE follows (
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);
CREATE INDEX idx_follows_following ON follows(following_id, created_at);
```

---

## 13. Фронтенд: план реализации (Phase F–K)

Фронтенд строится на Next.js 16 (App Router), React 19, TanStack Query, shadcn/ui. Архитектура — FSD-inspired: `entities/` → `features/` → `widgets/` → `views/`.

### Фаза F (4-6 дней): Backend bridge + Virtual scroll

**F1 — Пагинация пинов (backend bridge):**

Сделать `ListPins` cursor-based. Фронт переходит с `fetchPins()` на `fetchPinsPaginated(cursor?, limit?)`.

**F2 — User profile endpoints (backend bridge):**

Реализовать `GET /api/users/{id}` и `GET /api/users/{id}/pins`. На фронте — `src/entities/user/api/api.ts`: добавить `fetchUserById(id)`, `fetchUserPins(id, cursor?, limit?)`.

**F3 — Fix роутинга (backend bridge):**

- `router.go:45`: `POST /api/recommendations` → `GET /api/recommendations`
- `router.go`: добавить `DELETE /api/boards/{boardId}/pins/{pinId}`

**F4 — Infinite scroll (pin feed):**

- `src/entities/pin/api/api.ts`:
  - `useInfinitePins(initialData?)` — TanStack `useInfiniteQuery` с `getNextPageParam` (извлекает cursor из последнего пина страницы)
  - Параметр `?cursor=<created_at>` и `?limit=<int>` (default 50)
  - `staleTime: 5 * 60 * 1000`

- `src/widgets/pin-feed/ui/PinFeed.tsx` → рефакторинг:
  - `InfinitePinFeed` — client component с `IntersectionObserver`
  - Sentinel div внизу контейнера
  - `fetchNextPage()` при пересечении
  - Skeleton: 6 placeholder-блоков (серые анимированные прямоугольники с случайной высотой)
  - `isFetchingNextPage` — показывать скелетон внизу, **не** перекрывать весь контент

- `src/views/HomeView/ui/page.tsx`:
  - SSR: `fetchPinsPaginated({limit: 50})` → `initialPins`
  - Передать `initialPages` в `InfinitePinFeed`

**F5 — Обогащение Pin типа:**

- `src/entities/pin/model/types.ts`:
  ```ts
  export interface Pin {
    id: number;
    title: string;
    image_url: string;
    description: string;
    width: number;
    height: number;
    dominant_color: string;
    author?: { id: number; display_name: string; avatar_url: string };
    tags: string[];
    saves_count: number;
    is_saved: boolean;
    likes_count: number;
  }
  ```
- `src/entities/user/model/types.ts` — без изменений (User уже есть)
- `src/entities/board/model/types.ts` — новый тип Board

---

### Phase G (4-5 дней): Взаимодействия (save, like, comment, interaction tracking)

**G1 — Save/Unsave пина:**

- `src/features/save-pin/`:
  - `api.ts`: `useSavePin()` / `useUnsavePin()` — `useMutation` на `POST/DELETE /api/boards/{boardId}/pins/{pinId}`
  - `ui/SaveButton.tsx`:
    - На PinCard: появляется на hover, красное "Сохранить"
    - На PinPage: всегда видна, показывает выбранную доску
  - `ui/BoardSelectModal.tsx`:
    - Список `/api/boards/me`
    - Radio select + "Создать доску"
    - Если досок нет — сразу показать CreateBoardModal
  - `ui/CreateBoardModal.tsx`:
    - Поле title + checkbox is_private
    - `useCreateBoard()` mutation → POST /api/boards
    - После создания — автоматом сохранить пин

- `src/entities/board/`:
  - `model/types.ts`: Board { id, owner_id, title, is_private, created_at, pins_count? }
  - `api/api.ts`: fetchMyBoards, useMyBoards, createBoard, useCreateBoard

**G2 — Pin detail page (enhanced):**

- `src/views/PinPage/ui/page.tsx` — SSR:
  - Две колонки:
    - Левая: изображение (object-contain, max-h-[80vh])
    - Правая:
      - Save button (с BoardSelectModal)
      - Заголовок (h1), описание (p), теги (Chiplist)
      - **Автор**: аватар (48px) + display_name → Link `/users/[id]`
      - **Статистика**: saves_count + likes_count
      - **Like button**: иконка сердца, `useTrackInteraction('click', {action:'like'})`
  - `generateMetadata`: OG image, title, description

- `src/views/PinPage/ui/ClientComments.tsx`:
  - Dynamic import (ssr: false) — секция комментариев

**G3 — Comments:**

- `src/entities/comment/`:
  - `model/types.ts`: Comment { id, pin_id, author: User, body, created_at, updated_at }
  - `api/api.ts`: useComments(pinId), useCreateComment(pinId), useDeleteComment(commentId)

- `src/features/comment/ui/CommentList.tsx`:
  - `useComments(pinId)` — infinite query с cursor
  - Аватар + имя + текст + дата (relative time)
  - Delete button (только свои)
  - Empty state: "Пока нет комментариев"

- `src/features/comment/ui/CommentForm.tsx`:
  - Textarea + Submit button
  - `useCreateComment(pinId)` mutation
  - Optimistic update: добавить комментарий в список сразу

**G4 — Interaction tracking (frontend):**

- `src/shared/lib/tracker.ts`:
  - `Tracker` singleton (или React Context)
  - `track(eventType, pinId, meta?)` — POST /api/interactions
  - Session ID: `crypto.randomUUID()`, сохранить в sessionStorage
  - Client-side debounce: queue до 20 событий, flush каждые 5с или при отправке

- `src/features/tracker/`:
  - `useTrackImpression(ref)` — IntersectionObserver на PinCard → `track('impression', pinId)`
  - `useTrackClick(pinId)` — onClick → `track('click', pinId)`
  - `useTrackDwell(pinId)` — timeout 5s на PinPage → `track('dwell', pinId)`

- PinCard: прокинуть `useTrackImpression` (через ref)
- Pin detail: `useEffect` → `track('open_pin', pinId)` + `useTrackDwell(pinId)`

**G5 — Like tracking:**

- Like — это interaction с `event_type=click`, `meta={action:"like"}`
- `src/features/like/ui/LikeButton.tsx`:
  - Сердечко (filled/outline)
  - `track('click', pinId, {action: 'like'})` / `track('click', pinId, {action: 'unlike'})`
  - Optimistic update на saves_count

---

### Phase H (3-4 дня): Профиль пользователя + Follow

**H1 — Follow system (backend + frontend):**

- Backend: endpoint-ы из раздела 13.5
- `src/entities/follow/`:
  - `api/api.ts`: useFollow(userId), useUnfollow(userId), useFollowers(userId), useFollowing(userId)
- `src/features/follow/ui/FollowButton.tsx`:
  - "Подписаться" / "Отписаться"
  - `useFollow` / `useUnfollow` mutation
  - Invalidate query key `['user', userId]` после подписки

**H2 — User page (/users/[id]):**

- `app/users/[id]/page.tsx` → SSR
- `src/views/ProfileView/ui/page.tsx`:
  - Profile header:
    - Avatar (large, 120px)
    - Display name
    - Stats: N pins · N followers · N following
    - FollowButton (если не свой профиль)
    - Edit profile button (если свой профиль) — пока без реализации
  - Tabs: **Pins** | **Boards**
    - Pins tab: `InfinitePinFeed` с `fetchUserPins(userId)`
    - Boards tab: сетка досок (по 3 в ряд)
      - BoardCard: обложка (первый пин или placeholder), title, pin count
      - Link `/boards/[id]` (страницу доски делаем позже)

- `src/views/ProfileView/ui/ProfileHeader.tsx` — вынести header отдельно

**H3 — Header improvements:**

- `src/widgets/header/ui/Header.tsx`:
  - DropdownMenu: "Profile" → `/users/{currentUser.id}`, "Logout"
  - Search input: теперь рабочий — `onSubmit` → `router.push(/search?q=${query})`

**H4 — Board detail page (задел):**

- `app/boards/[id]/page.tsx`:
  - Board title
  - PinFeed (пины доски через `GET /api/boards/{boardId}/pins`)
  - Если доска приватная и не владелец — 403
  - **Отложить**, если не успеваем в MVP

---

### Phase I (2-3 дня): Поиск

**I1 — Search endpoint (backend):**

- Реализовать `GET /api/search?q=&cursor=&limit=` (раздел 13.6)

**I2 — Search page:**

- `app/search/page.tsx`:
  - Read `?q=` из searchParams
  - SSR: `fetchSearchResults(q, {limit: 50})`
  - `src/views/SearchView/ui/page.tsx`:
    - SearchResults — `InfinitePinFeed` с `useInfiniteSearchResults(q)`
    - Empty state: "Ничего не найдено для «{q}»"
    - Loading state: skeleton
    - Meta: `noindex`

- `src/entities/pin/api/api.ts` — add:
  - `fetchSearch(q, cursor?, limit?)`
  - `useInfiniteSearch(q)`

---

### Phase J (2-3 дня): UI Polish

**J1 — Состояния страниц:**

- **Loading skeletons:**
  - `PinFeed` → skeleton grid (10 блоков)
  - `PinPage` → 2-column layout skeleton
  - `ProfileView` → avatar circle + title skeleton + tabs skeleton
  - `SearchView` → skeleton grid
  - `BoardPage` → skeleton

- **Error boundaries:**
  - `src/shared/ErrorBoundary.tsx` — client component, catch `Error` from children
  - `ErrorFallback` — "Что-то пошло не так" + кнопка "Повторить"
  - Обернуть: PinFeed, PinPage, ProfileView, CreatePin

- **Empty states:**
  - Feed: "Лента пуста. Подпишитесь на пользователей или создайте первый пин!" + CTA
  - Profile pins: "Пользователь пока не создал пины"
  - Profile boards: "Доски не созданы"
  - Search: "Ничего не найдено"
  - Comments: "Пока нет комментариев. Напишите первый!"

**J2 — Mobile responsive:**

- `Header`:
  - На мобильных: гамбургер-меню (Drawer или Bottom sheet)
  - Search: expandable search bar (иконка → полный поиск)
- Bottom navigation bar:
  - Home, Search, Create, Profile
  - Fixed bottom, скрывается при скролле вниз
- PinFeed:
  - Mobile: `columns-2` всегда, отступы меньше
- PinPage:
  - Mobile: одна колонка (изображение сверху, контент снизу)

**J3 — Dark mode:**

- Проверить все страницы:
  - PinPage: белый фон карточки → dark:bg-gray-900
  - Header: glass-эффекты → dark:bg-gray-900/80
  - PinFeed skeleton → dark:bg-gray-800
  - Комментарии: фон → dark:bg-gray-800

**J4 — Анимации:**

- PinCard:
  - hover: `scale-[1.02]` + `shadow-lg` (transition 200ms)
  - overlay: title появляется снизу
- Save button:
  - Появление: `opacity-0 → opacity-100` на hover карточки
  - Pulse при сохранении
- Like:
  - `scale-125` на 150ms при клике (heartbeat)

---

### Phase K (1-2 дня): Auth & Protected routes refinement

**K1 — Auth guard middleware:**

- `proxy.ts`: расширить matcher
  ```
  matcher: ['/create/:path*', '/profile/:path*', '/boards/:path*']
  ```
- Редирект: `/signin?from=${pathname}`

**K2 — Client-side auth state:**

- `src/features/auth/useAuth.ts`:
  - Хранить `user` в React Query (`['user', 'currentUser']`)
  - `session_id` в localStorage для interaction tracker
  - При 401 → `queryClient.setQueryData(['user', 'currentUser'], null)` — не редиректить, просто показать анонимный UI

**K3 — 401 handling:**

- В `src/shared/lib/api.ts` — общий fetch wrapper:
  - Если ответ 401 — `invalidateQueries(['user', 'currentUser'])`
  - Не редиректить принудительно, дать UI самому решить

---

### Порядок реализации (recommended sequence)

| Порядок | Фаза    | Зависит от                                             |
| ------- | ------- | ------------------------------------------------------ |
| 1       | F1–F3   | Backend bridge: pagination, user endpoints, fix router |
| 2       | F4–F5   | Infinite scroll + enriched Pin type                    |
| 3       | G1      | Save/Unsave + BoardSelectModal                         |
| 4       | G4      | Interaction tracking (нужен для лайков)                |
| 5       | G5 + G2 | Like + Pin detail page enhancement                     |
| 6       | G3      | Comments (нужна PinPage)                               |
| 7       | H1–H3   | Follow + Profile page                                  |
| 8       | I1–I2   | Search                                                 |
| 9       | J1–J4   | Polish                                                 |
| 10      | K1–K3   | Auth refinement                                        |

---

## 15. Definition of Done по этапам (фронтенд)

Готово, если:

- [ ] Infinite scroll работает: подгрузка при скролле, skeleton, не сбрасывает позицию
- [ ] PinCard показывает автора, saves_count, кнопку save на hover
- [ ] Save/unsave через BoardSelectModal корректно работает
- [ ] PinPage: автор, save, related pins, комментарии
- [ ] Interaction tracking отправляет события без потери
- [ ] /users/[id] показывает профиль + пины + доски + follow button
- [ ] Follow/unfollow работает с оптимистичным обновлением
- [ ] /search?q= находит пины по названию и тегам
- [ ] Все страницы имеют skeleton загрузки и empty state
- [ ] Мобильная версия адекватна (bottom nav, columns-2)
- [ ] Protected routes редиректят на /signin
- [ ] `npm run lint` и `npm run build` проходят без ошибок
