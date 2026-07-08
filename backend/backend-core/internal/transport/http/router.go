package httpserver

import (
	"net/http"

	"pinterest-core/internal/transport/http/handlers"
	"pinterest-core/internal/transport/http/middleware"
)

func NewRouter(pinHandler *handlers.PinHandler, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/pins", pinHandler.ListPins)
	mux.HandleFunc("/api/pins/{id}", pinHandler.GetPinByID)
	mux.HandleFunc("/api/pins/create", pinHandler.CreatePin)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/users", userHandler.ListUsers)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/user/current", authHandler.Current)
	return middleware.CorsMiddleware(mux)
}
