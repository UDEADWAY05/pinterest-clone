package handlers

import (
	"net/http"
	"pinterest-core/internal/usecase"
	"strconv"

	"pinterest-core/internal/transport/http/render"
)

type UserHandler struct {
    UC *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
    return &UserHandler{UC: uc}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.UC.List(r.Context())
    if err != nil {
        render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load users", "")
        return
    }

    render.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is required", "")
        return
    }

    idInt, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is not valid", "")
        return
    }

    user, err := h.UC.GetByID(r.Context(), idInt)
    if err != nil {
        render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user", "")
        return
    }

    render.JSON(w, http.StatusOK, user)
}
