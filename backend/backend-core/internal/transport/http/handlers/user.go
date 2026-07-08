package handlers

import (
	"encoding/json"
	"net/http"
	"pinterest-core/internal/usecase"
	"strconv"
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
		 http.Error(w, "failed to load users", http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	idInt, idIntErr := strconv.ParseInt(id, 10, 64)

	if idIntErr != nil {
	 http.Error(w, "id is not valid", http.StatusBadRequest)
		return
	}

	pin, err := h.UC.GetByID(r.Context(), idInt)

	if err != nil {
		 http.Error(w, "failed to get pin", http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pin)
}
