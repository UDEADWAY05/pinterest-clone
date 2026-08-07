package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

type FollowHandler struct {
	UC *usecase.FollowUsecase
}

func NewFollowHandler(uc *usecase.FollowUsecase) *FollowHandler {
	return &FollowHandler{UC: uc}
}

// FollowUser подписывает текущего пользователя на указанного.
// @Summary      Подписаться
// @Description  Текущий авторизованный пользователь подписывается на пользователя с id
// @Tags         follow
// @Param        id  path  int  true  "ID пользователя, на которого подписываются"
// @Success      200  {object}  render.JSONResponse{data=map[string]string}
// @Failure      400  {object}  render.ErrorResponse
// @Failure      401  {object}  render.ErrorResponse
// @Failure      404  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/users/{id}/follow [post]
func (h *FollowHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid session", "")
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || targetID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	if err := h.UC.Follow(r.Context(), user.ID, targetID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			render.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found", "")
			return
		}
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	render.JSON(w, http.StatusOK, map[string]bool{"following": true})
}

// UnfollowUser отписывает текущего пользователя.
// @Summary      Отписаться
// @Description  Текущий авторизованный пользователь отписывается от пользователя с id
// @Tags         follow
// @Param        id  path  int  true  "ID пользователя, от которого отписываются"
// @Success      200  {object}  render.JSONResponse{data=map[string]bool}
// @Failure      400  {object}  render.ErrorResponse
// @Failure      401  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/users/{id}/follow [delete]
func (h *FollowHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid session", "")
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || targetID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	if err := h.UC.Unfollow(r.Context(), user.ID, targetID); err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to unfollow", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, map[string]bool{"following": false})
}

// ListFollowers возвращает подписчиков пользователя.
// @Summary      Подписчики
// @Description  Пагинированный список пользователей, подписанных на пользователя с id
// @Tags         follow
// @Produce      json
// @Param        id      path  int    true  "ID пользователя"
// @Param        cursor  query string false "Курсор (RFC3339Nano created_at)"
// @Param        limit   query int    false "Лимит (default 50)"
// @Success      200  {object}  render.JSONResponse{data=domain.ListFollowsResult}
// @Router       /api/users/{id}/followers [get]
func (h *FollowHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || userID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if n, parseErr := strconv.Atoi(limitStr); parseErr == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, cursorStr); parseErr == nil {
			cursor = &parsed
		}
	}

	result, err := h.UC.ListFollowers(r.Context(), userID, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load followers", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, result)
}

// ListFollowing возвращает подписки пользователя.
// @Summary      Подписки
// @Description  Пагинированный список пользователей, на которых подписан пользователь с id
// @Tags         follow
// @Produce      json
// @Param        id      path  int    true  "ID пользователя"
// @Param        cursor  query string false "Курсор (RFC3339Nano created_at)"
// @Param        limit   query int    false "Лимит (default 50)"
// @Success      200  {object}  render.JSONResponse{data=domain.ListFollowsResult}
// @Router       /api/users/{id}/following [get]
func (h *FollowHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || userID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user id", "")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if n, parseErr := strconv.Atoi(limitStr); parseErr == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, cursorStr); parseErr == nil {
			cursor = &parsed
		}
	}

	result, err := h.UC.ListFollowing(r.Context(), userID, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load following", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, result)
}
