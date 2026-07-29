# AGENTS.md

## Project overview

Pinterest clone: Go backend (Clean Architecture) + Next.js 16 frontend (App Router / FSD-inspired). ML worker (Python) planned but not integrated yet.

## Directory map

```
backend/backend-core/   # Go module "pinterest-core" — main API server
backend/ml-worker/       # Python ML service (planned, not wired in)
frontend/                # Next.js 16 SPA (React 19, Tailwind CSS 4, shadcn/ui)
seed.ts                  # TypeScript seeder — POSTs demo pins to backend
docker-compose.yml       # Root infra: pgvector + Redis + MinIO
```

## Infrastructure

- **PostgreSQL** (pgvector) on `:5432`, db `udpinterest`, user/pass `postgres/postgres`
- **Redis** on `:6379`
- **MinIO** (S3) on `:9000` (API) / `:9001` (console), bucket `pinterest-media`, creds `admin/supersecretpassword123`
- Start: `docker-compose up -d` (from repo root)
- The root `docker-compose.yml` (pgvector) is the canonical infra file. `backend/backend-core/docker-compose.yml` uses plain postgres — ignore it.

## Commands

### Backend (`backend/backend-core/`)

```bash
go run ./cmd/app          # start server (needs .env in backend/backend-core/)
air                       # hot reload (uses .air.toml)
go test ./...             # run all tests
```

### Frontend (`frontend/`)

```bash
npm run dev               # next dev on :3000
npm run build             # production build
npm run lint              # ESLint (next/core-web-vitals + typescript rules)
```

### Seeding

```bash
npx tsx seed.ts           # run from repo root — POSTs hardcoded pins to backend
```

## Key architecture facts

### Backend (Clean Architecture)

- Entrypoint: `cmd/app/main.go` — wires config, DB, Redis, S3, repos, usecases, handlers
- Layers: `transport/http/` → `usecase/` → `repository/` → `domain/`
- All routes defined in `internal/transport/http/router.go:24`
- Auth: cookie-based sessions (Redis-backed), `session` cookie. Middleware at `middleware/auth_middleware.go`
- CORS: `middleware/cors_middleware.go`
- Config: `config/config.go` reads from env (godotenv loads `.env` from cwd)
- Migrations: `golang-migrate` auto-runs from `migrations/` dir on startup
- Swagger at `/swagger/`
- Module name is `pinterest-core` — imports use that prefix

### Frontend (FSD-inspired)

- `src/app/` — Next.js App Router pages
- `src/entities/pin/`, `src/entities/user/` — domain types/API clients
- `src/features/auth/` — auth feature
- `src/views/` — page-level views (HomeView, PinPage, etc.)
- `src/widgets/` — composite widgets (header, pin-feed)
- `src/shared/ui/` — shadcn/ui components
- `@/*` alias → `./src/*` (NOT `./*`)
- API calls proxy through Next.js rewrites: `/api/:path*` → `http://localhost:8080/api/:path*`
- Auth guard: `frontend/proxy.ts` redirects `/create/:path*` to `/signin` if no session cookie

## Gotchas

- Backend `.env` must be in `backend/backend-core/` (godotenv loads relative to cwd)
- MinIO bucket (`pinterest-media`) may need manual creation on first run
- `seed.ts` uses a hardcoded session cookie — run after creating a user/session
- Two docker-compose files exist; use the **root** one (has pgvector)
- Go module name is `pinterest-core` — don't use `backend-core` in import paths
- Frontend `tsconfig.json` paths alias `@/*` → `src/*`, not root
