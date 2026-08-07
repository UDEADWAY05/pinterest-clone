package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

type CommentHandler struct {
	UC *usecase.CommentUsecase
}

func NewCommentHandler(uc *usecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{UC: uc}
}

// GetComments возвращает комментарии пина, отсортированные по created_at ASC.
// @Summary      Комментарии пина
// @Description  Возвращает пагинированный список комментариев пина
// @Tags         comments
// @Produce      json
// @Param        id      path  int    true  "ID пина"
// @Param        cursor  query string false "Курсор (RFC3339Nano created_at последнего комментария)"
// @Param        limit   query int    false "Лимит (default 50)"
// @Success      200     {object}  render.JSONResponse{data=domain.ListCommentsResult}
// @Router       /api/pins/{id}/comments [get]
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	pinID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || pinID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pin id", "")
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

	result, err := h.UC.ListByPin(r.Context(), pinID, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load comments", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, result)
}

type createCommentRequest struct {
	Body string `json:"body"`
}

// CreateComment добавляет комментарий к пину.
// @Summary      Добавить комментарий
// @Description  Добавляет комментарий к пину от авторизованного пользователя
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "ID пина"
// @Param        body body  createCommentRequest  true  "Тело комментария"
// @Success      201  {object}  render.JSONResponse{data=domain.Comment}
// @Failure      400  {object}  render.ErrorResponse
// @Failure      401  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/pins/{id}/comments [post]
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid session", "")
		return
	}

	pinID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || pinID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pin id", "")
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", err.Error())
		return
	}

	comment, err := h.UC.Create(r.Context(), pinID, user.ID, req.Body)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	render.JSON(w, http.StatusCreated, comment)
}

// DeleteComment удаляет комментарий (только владелец).
// @Summary      Удалить комментарий
// @Description  Мягко удаляет комментарий; доступно только автору
// @Tags         comments
// @Param        id  path  int  true  "ID комментария"
// @Success      204  {object}  render.JSONResponse
// @Failure      401  {object}  render.ErrorResponse
// @Failure      404  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*domain.User)
	if !ok {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid session", "")
		return
	}

	commentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || commentID < 1 {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id", "")
		return
	}

	if err := h.UC.Delete(r.Context(), commentID, user.ID); err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			render.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found", "")
			return
		}
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete comment", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
