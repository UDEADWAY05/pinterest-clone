package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"pinterest-core/internal/config"
	"pinterest-core/internal/infra/db"
	"pinterest-core/internal/infra/redis"
	"pinterest-core/internal/repository"
	"pinterest-core/internal/repository/postgresrepo"
	"pinterest-core/internal/repository/redisrepo"
	httpserver "pinterest-core/internal/transport/http"
	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pg, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = pg.Close(cctx)
	}()

	redclient, err := redis.New(ctx, cfg)

	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()
		_ = redclient.Close(cctx)
	}()

	// ── Репозитории (без кеша) ──────────────────────────────────
	// Это «сырые» Postgres-реализации — прямой доступ к БД.
	// Мы не используем их напрямую в usecase'ах — оборачиваем в кеш-декораторы.
	pinPgRepo := postgresrepo.NewPinRepository(pg.Pool)
	userPgRepo := postgresrepo.NewUserRepository(pg.Pool)
	sessionRepo := redisrepo.NewSessionRepository(redclient.Client)

	// ── Кеш-слой ────────────────────────────────────────────────
	// CacheRepository — абстракция. Сейчас это Redis, но может быть
	// заменён на in-memory (для тестов) или другой бэкенд.
	cacheRepo := redisrepo.NewCacheRepository(redclient.Client)

	// ── Репозитории (с кешем) ───────────────────────────────────
	// Decorator Pattern: оборачиваем Postgres-репозитории в кеширующие.
	// UseCase'ы ничего не знают о кеше — они работают с интерфейсом.
	pinRepo := repository.NewCachedPinRepository(
		pinPgRepo, cacheRepo,
		cfg.PinCacheTTL(),  // TTL одного пина: 5 мин (CACHE_PIN_TTL)
		cfg.ListCacheTTL(), // TTL списка: 1 мин (CACHE_LIST_TTL)
	)
	userRepo := repository.NewCachedUserRepository(
		userPgRepo, cacheRepo,
		cfg.UserCacheTTL(), // TTL профиля: 10 мин (CACHE_USER_TTL)
	)

	// ── UseCase'ы ───────────────────────────────────────────────
	// Передаём уже закешированные репозитории.
	// UseCase не знает, откуда пришли данные — из кеша или из БД.
	pinUC := usecase.NewPinUsecase(pinRepo)
	authUC := usecase.NewAuthUsecase(userRepo, sessionRepo, 7*24*time.Hour)
	userUC := usecase.NewUserUsecase(userRepo)

	pinHandler := handlers.NewPinHandler(pinUC, authUC)
	authHandler := handlers.NewAuthHandler(authUC)
	userHandlers := handlers.NewUserHandler(userUC)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", "./migrations"),
		cfg.PostgresDSN,
	)
	if err != nil {
		log.Fatalf("failed to initialize database migration: %v", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("database migration failed: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: httpserver.NewRouter(pinHandler, authHandler, userHandlers),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}
}
