package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pinterest-core/internal/config"
	"pinterest-core/internal/infra/db"
	"pinterest-core/internal/infra/redis"
	"pinterest-core/internal/repository/postgresrepo"
	"pinterest-core/internal/repository/redisrepo"
	httpserver "pinterest-core/internal/transport/http"
	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/usecase"

	"github.com/joho/godotenv"
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

	pinRepo := postgresrepo.NewPinRepository(pg.Pool)
	userRepo := postgresrepo.NewUserRepository(pg.Pool)
	sessionRepo := redisrepo.NewSessionRepository(redclient.Client)

	pinUC := usecase.NewPinUsecase(pinRepo)
	authUC := usecase.NewAuthUsecase(userRepo, sessionRepo, 7*24*time.Hour)
	userUC := usecase.NewUserUsecase(userRepo)

	pinHandler := handlers.NewPinHandler(pinUC, authUC)
	authHandler := handlers.NewAuthHandler(authUC)
	userHandlers := handlers.NewUserHandler(userUC)
	
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
