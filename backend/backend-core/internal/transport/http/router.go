package httpserver

import (
	"net/http"

	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/transport/http/middleware"
)

func NewRouter(
	pinHandler *handlers.PinHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	boardHandler *handlers.BoardHandler,
	uploadHandler *handlers.UploadHandler,
	interactionHandler *handlers.InteractionHandler,
	recommendationHandler *handlers.RecommendationHandler,
) http.Handler {
	mux := http.NewServeMux()

	// ── Public routes ──────────────────────────────────────────────
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/pins", pinHandler.ListPins)
	mux.HandleFunc("/api/pins/{id}", pinHandler.GetPinByID)
	mux.HandleFunc("/api/pins/{id}/related", recommendationHandler.GetRelated)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/users", userHandler.ListUsers)
	mux.HandleFunc("/api/boards/{boardId}/pins", boardHandler.GetBoardPins)
	mux.HandleFunc("/api/interactions", interactionHandler.Save)
	mux.HandleFunc("/api/recommendations", recommendationHandler.GetFeed)

	// ── Auth middleware ────────────────────────────────────────────
	authMw := middleware.AuthMiddleware(authHandler.UC)

	// ── Protected routes (auth required) ───────────────────────────
	mux.HandleFunc("/api/pins/create", authMw(pinHandler.CreatePin))
	mux.HandleFunc("/api/auth/logout", authMw(authHandler.Logout))
	mux.HandleFunc("/api/user/current", authMw(authHandler.Current))
	mux.HandleFunc("/api/boards/me", authMw(boardHandler.ListMyBoards))
	mux.HandleFunc("/api/boards", authMw(boardHandler.CreateBoard))
	mux.HandleFunc("/api/boards/{boardId}/pins/{pinId}", authMw(boardHandler.HandlePinInBoard))
	mux.HandleFunc("/api/uploads/presign", authMw(uploadHandler.PresignUpload))

	return middleware.CorsMiddleware(mux)
}
