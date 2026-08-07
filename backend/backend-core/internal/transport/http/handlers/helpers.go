package handlers

import (
	"net/http"

	"pinterest-core/internal/usecase"
)

// currentViewerID возвращает ID авторизованного пользователя из session cookie,
// либо nil, если пользователь не авторизован.
func currentViewerID(r *http.Request, authUC *usecase.AuthUsecase) *int64 {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	user, err := authUC.Current(r.Context(), cookie.Value)
	if err != nil {
		return nil
	}
	return &user.ID
}
