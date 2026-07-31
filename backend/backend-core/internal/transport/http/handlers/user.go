package handlers

import (
	"errors"
	"net/http"
	"pinterest-core/internal/repository"
	"pinterest-core/internal/usecase"
	"strconv"
	"time"

	"pinterest-core/internal/transport/http/render"
)

type UserHandler struct {
	UC     *usecase.UserUsecase
	PinUC  *usecase.PinUsecase
	AuthUC *usecase.AuthUsecase
}

func NewUserHandler(uc *usecase.UserUsecase, pinUC *usecase.PinUsecase, authUC *usecase.AuthUsecase) *UserHandler {
	return &UserHandler{UC: uc, PinUC: pinUC, AuthUC: authUC}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UC.List(r.Context())
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load users", "")
		return
	}

	render.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	var currentUserID *int64
	cookie, cookieErr := r.Cookie("session")
	if cookieErr == nil {
		user, authErr := h.AuthUC.Current(r.Context(), cookie.Value)
		if authErr == nil {
			currentUserID = &user.ID
		}
	}

	profile, err := h.UC.GetProfile(r.Context(), id, currentUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			render.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found", "")
			return
		}
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load profile", "")
		return
	}

	render.JSON(w, http.StatusOK, profile)
}

func (h *UserHandler) ListUserCreatedPins(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		n, parseErr := strconv.Atoi(limitStr)
		if parseErr == nil && n >= 1 {
			limit = n
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, cursorStr)
		if parseErr == nil {
			cursor = &parsed
		}
	}

	result, err := h.PinUC.ListUserCreatedPins(r.Context(), userID, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", "")
		return
	}

	render.JSON(w, http.StatusOK, result)
}

func (h *UserHandler) ListUserSavedPins(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		n, parseErr := strconv.Atoi(limitStr)
		if parseErr == nil && n >= 1 {
			limit = n
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, cursorStr)
		if parseErr == nil {
			cursor = &parsed
		}
	}

	result, err := h.PinUC.ListUserSavedPins(r.Context(), userID, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", "")
		return
	}

	render.JSON(w, http.StatusOK, result)
}

func (h *UserHandler) getCurrentUserID(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, errors.New("missing session token")
	}

	user, err := h.AuthUC.Current(r.Context(), cookie.Value)
	if err != nil {
		return 0, errors.New("invalid or expired session")
	}

	return user.ID, nil
}
