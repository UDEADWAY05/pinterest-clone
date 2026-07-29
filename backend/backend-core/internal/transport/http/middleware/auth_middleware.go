package middleware

import (
	"context"
	"net/http"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// AuthMiddleware verifies that a session token exists and is valid before calling the next handler.
func AuthMiddleware(authUC *usecase.AuthUsecase) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("session")
            if err != nil || cookie.Value == "" {
                render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing session token", "")
                return
            }

            // Validate the session token; we don't need to use the returned user here.
            user, err := authUC.Current(r.Context(), cookie.Value)
            if err != nil {
                render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid session", "")
                return
            }

            // Set the user in the context for use in the next handler.
            ctx := r.Context()
            ctx = context.WithValue(ctx, "user", user)
            r = r.WithContext(ctx)

            next.ServeHTTP(w, r)
        })
    }
}
