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

	"pinterest-core/internal/config"
	"pinterest-core/internal/infra/db"
	"pinterest-core/internal/infra/redis"
	"pinterest-core/internal/infra/storage"
	"pinterest-core/internal/repository"
	"pinterest-core/internal/repository/postgresrepo"
	"pinterest-core/internal/repository/redisrepo"
	"pinterest-core/internal/repository/storagerepo"
	httpserver "pinterest-core/internal/transport/http"
	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/usecase"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// ── PostgreSQL ───────────────────────────────────────────────────
	pg, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = pg.Close(cctx)
	}()

	// ── Redis ───────────────────────────────────────────────────────
	redclient, err := redis.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = redclient.Close(cctx)
	}()

	// ── S3 / MinIO (опционально) ────────────────────────────────────
	// Если S3_ENDPOINT не задан — storage.New вернёт ошибку, пропускаем.
	s3Client, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to S3: %v", err)
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s3Client.Close(cctx)
	}()


	// ── Репозитории (без кеша) ──────────────────────────────────────
	pinPgRepo := postgresrepo.NewPinRepository(pg.Pool)
	userPgRepo := postgresrepo.NewUserRepository(pg.Pool)
	boardPgRepo := postgresrepo.NewBoardRepository(pg.Pool)
	sessionRepo := redisrepo.NewSessionRepository(redclient.Client)

	// ── Кеш-слой ────────────────────────────────────────────────────
	cacheRepo := redisrepo.NewCacheRepository(redclient.Client)

	// ── Репозитории (с кешем) ───────────────────────────────────────
	pinRepo := repository.NewCachedPinRepository(pinPgRepo, cacheRepo, cfg.PinCacheTTL(), cfg.ListCacheTTL())
	userRepo := repository.NewCachedUserRepository(userPgRepo, cacheRepo, cfg.UserCacheTTL())
	boardRepo := repository.NewCachedBoardRepository(boardPgRepo, cacheRepo, cfg.ListCacheTTL(), cfg.BoardCacheTTL())

	// ── UseCase'ы ───────────────────────────────────────────────────
	pinUC := usecase.NewPinUsecase(pinRepo)
	authUC := usecase.NewAuthUsecase(userRepo, sessionRepo, 7*24*time.Hour)
	userUC := usecase.NewUserUsecase(userRepo)
	boardUC := usecase.NewBoardUsecase(boardRepo, pinRepo)

	// ── Storage (S3, если доступен) + Upload ───────────────────────
	var uploadHandler *handlers.UploadHandler
	publicBaseURL := cfg.S3PublicURL
	if publicBaseURL == "" {
		scheme := "http://"
		if cfg.S3UseSSL {
			scheme = "https://"
		}
		publicBaseURL = scheme + cfg.S3Endpoint
	}

	storageRepo := storagerepo.NewMinioStorage(s3Client.Client, cfg.S3Bucket, publicBaseURL)
	uploadUC := usecase.NewUploadUsecase(storageRepo, cacheRepo, cfg.PresignExpiry())
	uploadHandler = handlers.NewUploadHandler(uploadUC, authUC)


	// ── Interactions (сбор событий) ─────────────────────────────────
	interactionRepo := postgresrepo.NewInteractionRepository(pg.Pool)
	interactionUC := usecase.NewInteractionUsecase(interactionRepo, 50, 3*time.Second)
	interactionHandler := handlers.NewInteractionHandler(interactionUC)

	// ── Recommendations ─────────────────────────────────────────────
	recommendationUC := usecase.NewRecommendationUsecase(pinRepo, cacheRepo)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationUC, authUC)

	// ── Миграции ────────────────────────────────────────────────────
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

	// ── Handlers ────────────────────────────────────────────────────
	pinHandler := handlers.NewPinHandler(pinUC)
	authHandler := handlers.NewAuthHandler(authUC)
	userHandlers := handlers.NewUserHandler(userUC)
	boardHandler := handlers.NewBoardHandler(boardUC, authUC)

	// ── HTTP Server ─────────────────────────────────────────────────
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: httpserver.NewRouter(pinHandler, authHandler, userHandlers, boardHandler, uploadHandler, interactionHandler, recommendationHandler),
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

	// ── Shutdown ────────────────────────────────────────────────────
	// Сначала сбрасываем буфер событий, потом выключаем HTTP
	interactionUC.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}
}
