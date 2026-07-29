package handlers

import (
	"net/http"
	"pinterest-core/internal/usecase"

	"pinterest-core/internal/transport/http/render"
)

type UserHandler struct {
    UC *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
    return &UserHandler{UC: uc}
}

// ListUsers возвращает список пользователей.
// @Summary      Список пользователей
// @Description  Возвращает всех пользователей
// @Tags         users
// @Produce      json
// @Success      200  {object}  render.JSONResponse{data=[]domain.User}
// @Router       /api/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.UC.List(r.Context())
    if err != nil {
        render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load users", "")
        return
    }

    render.JSON(w, http.StatusOK, users)
}
