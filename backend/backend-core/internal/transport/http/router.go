package httpserver

import (
	"net/http"

	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/transport/http/middleware"

	_ "pinterest-core/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Pinterest Core API
// @version         1.0
// @description     API backend-сервиса Pinterest.
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func NewRouter(
	cfgCors middleware.CORSConfig,
	pinHandler *handlers.PinHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	boardHandler *handlers.BoardHandler,
	uploadHandler *handlers.UploadHandler,
	interactionHandler *handlers.InteractionHandler,
	recommendationHandler *handlers.RecommendationHandler,
	commentHandler *handlers.CommentHandler,
	followHandler *handlers.FollowHandler,
) http.Handler {
	mux := http.NewServeMux()

	// ── Public routes ──────────────────────────────────────────────
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /api/pins", pinHandler.ListPins)
	mux.HandleFunc("GET /api/search", pinHandler.SearchPins)
	mux.HandleFunc("GET /api/pins/{id}", pinHandler.GetPinByID)
	mux.HandleFunc("GET /api/pins/{id}/related", recommendationHandler.GetRelated)
	mux.HandleFunc("GET /api/pins/{id}/comments", commentHandler.GetComments)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/users", userHandler.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetProfile)
	mux.HandleFunc("GET /api/users/{id}/created", userHandler.ListUserCreatedPins)
	mux.HandleFunc("GET /api/users/{id}/saved", userHandler.ListUserSavedPins)
	mux.HandleFunc("GET /api/users/{id}/followers", followHandler.ListFollowers)
	mux.HandleFunc("GET /api/users/{id}/following", followHandler.ListFollowing)
	mux.HandleFunc("GET /api/users/{id}/boards", boardHandler.GetUserBoards)
	mux.HandleFunc("GET /api/boards/{boardId}/pins", boardHandler.GetBoardPins)
	mux.HandleFunc("POST /api/interactions", interactionHandler.Save)
	mux.HandleFunc("GET /api/recommendations", recommendationHandler.GetFeed)

	// ── Auth middleware ────────────────────────────────────────────
	authMw := middleware.AuthMiddleware(authHandler.UC)

	// ── Protected routes (auth required) ───────────────────────────
	mux.HandleFunc("POST /api/pins/create", authMw(pinHandler.CreatePin))
	mux.HandleFunc("POST /api/auth/logout", authMw(authHandler.Logout))
	mux.HandleFunc("GET /api/user/current", authMw(authHandler.Current))
	mux.HandleFunc("GET /api/boards/me", authMw(boardHandler.ListMyBoards))
	mux.HandleFunc("POST /api/boards", authMw(boardHandler.CreateBoard))
	mux.HandleFunc("DELETE /api/boards", authMw(boardHandler.RemoveBoard))
	mux.HandleFunc("POST /api/boards/{boardId}/pins/{pinId}", authMw(boardHandler.SavePinToBoard))
	mux.HandleFunc("DELETE /api/boards/{boardId}/pins/{pinId}", authMw(boardHandler.RemovePinFromBoard))
	mux.HandleFunc("POST /api/uploads/presign", authMw(uploadHandler.PresignUpload))
	mux.HandleFunc("PATCH /api/pins/{id}", authMw(pinHandler.UpdatePin))
	mux.HandleFunc("DELETE /api/pins/{id}", authMw(pinHandler.DeletePin))
	mux.HandleFunc("POST /api/pins/{id}/comments", authMw(commentHandler.CreateComment))
	mux.HandleFunc("DELETE /api/comments/{id}", authMw(commentHandler.DeleteComment))
	mux.HandleFunc("POST /api/users/{id}/follow", authMw(followHandler.FollowUser))
	mux.HandleFunc("DELETE /api/users/{id}/follow", authMw(followHandler.UnfollowUser))

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	return middleware.CorsMiddleware(cfgCors, mux)
}
