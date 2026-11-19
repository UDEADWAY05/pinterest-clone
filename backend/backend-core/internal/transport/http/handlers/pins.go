package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"pinterest-core/internal/domain"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
	"strconv"
	"strings"
	"time"
)

type PinHandler struct {
	UC     *usecase.PinUsecase
	AuthUC *usecase.AuthUsecase
}

func NewPinHandler(uc *usecase.PinUsecase, authUC *usecase.AuthUsecase) *PinHandler {
	return &PinHandler{UC: uc, AuthUC: authUC}
}

// ListPins возвращает список всех пинов.
// @Summary      Список пинов
// @Description  Возвращает все пины, отсортированные по дате создания
// @Tags         pins
// @Produce      json
// @Success      200  {object}  render.JSONResponse{data=domain.ListPinsResult}
// @Router       /api/pins [get]
func (h *PinHandler) ListPins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n < 1 {
			limit = 20
		} else {
			limit = n
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	var cursor *time.Time

	if cursorStr != "" {
		parsedCursor, err := time.Parse(time.RFC3339Nano, cursorStr)

		if err != nil {
			cursor = nil
		} else {
			cursor = &parsedCursor
		}
	}

	var userId *int64
	if userIdStr := r.URL.Query().Get("user_id"); userIdStr != "" {
		if id, err := strconv.ParseInt(userIdStr, 10, 64); err == nil && id > 0 {
        userId = &id
    }
	}

	pins, err := h.UC.List(r.Context(), cursor, limit, userId)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", err.Error())
		return
	}

	pins.Data, err = h.UC.EnrichPins(r.Context(), pins.Data, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

// SearchPins возвращает список пинов по поисковому запросу.
// @Summary      Поиск пинов
// @Description  Ищет пины по названию и тегам (ILIKE по pins.title и pin_tags.tag)
// @Tags         pins
// @Produce      json
// @Param        q       query string true  "Поисковый запрос"
// @Param        cursor  query string false "Курсор (RFC3339Nano created_at)"
// @Param        limit   query int    false "Лимит (default 50)"
// @Success      200  {object}  render.JSONResponse{data=domain.ListPinsResult}
// @Router       /api/search [get]
func (h *PinHandler) SearchPins(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "q is required", "")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n < 1 {
			limit = 50
		} else {
			limit = n
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, cursorStr); err == nil {
			cursor = &parsed
		}
	}

	pins, err := h.UC.Search(r.Context(), q, cursor, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search pins", err.Error())
		return
	}

	pins.Data, err = h.UC.EnrichPins(r.Context(), pins.Data, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search pins", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

// GetPinByID возвращает один пин по ID.
// @Summary      Получить пин
// @Description  Возвращает детальную информацию о пине по его ID
// @Tags         pins
// @Produce      json
// @Param        id   path      int  true  "ID пина"
// @Success      200  {object}  render.JSONResponse{data=domain.Pin}
// @Failure      400  {object}  render.ErrorResponse
// @Router       /api/pins/{id} [get]
func (h *PinHandler) GetPinByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is required", "")
		return
	}

	idInt, idIntErr := strconv.ParseInt(id, 10, 64)
	if idIntErr != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is not valid", idIntErr.Error())
		return
	}

	pin, err := h.UC.GetByID(r.Context(), idInt)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, http.StatusNotFound, "NOT_FOUND", "pin not found", err.Error())
			return
		}
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get pin", err.Error())
		return
	}

	enriched, err := h.UC.EnrichPins(r.Context(), []domain.Pin{*pin}, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get pin", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, enriched[0])
}

type CreatePinRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	ImageURL      string   `json:"image_url"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	DominantColor string   `json:"dominant_color"`
	Tags          []string `json:"tags,omitempty"`
}

// CreatePin создаёт новый пин.
// @Summary      Создать пин
// @Description  Создаёт новый пин. Требует авторизации (session cookie)
// @Tags         pins
// @Accept       json
// @Produce      json
// @Param        body  body      CreatePinRequest  true  "Данные пина"
// @Success      201  {object}  render.JSONResponse{data=domain.Pin}
// @Failure      400  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/pins/create [post]
func (h *PinHandler) CreatePin(w http.ResponseWriter, r *http.Request) {
	var req CreatePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	user := r.Context().Value("user").(*domain.User)
	authorID := user.ID

	pin, err := h.UC.Create(r.Context(), req.Title, req.Description, req.DominantColor, req.ImageURL, req.Width, req.Height, authorID, req.Tags)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			render.Error(w, http.StatusNotFound, "NOT_FOUND", "pin not found", err.Error())
			return
		}
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get pin", err.Error())
		return
	}
	render.JSON(w, http.StatusCreated, pin)
}

type UpdatePinRequest struct {
	Title         *string   `json:"title,omitempty"`
	Description   *string   `json:"description,omitempty"`
	ImageURL      *string   `json:"image_url,omitempty"`
	Width         *int      `json:"width,omitempty"`
	Height        *int      `json:"height,omitempty"`
	DominantColor *string   `json:"dominant_color,omitempty"`
	Tags          *[]string `json:"tags,omitempty"`
}

func (h *PinHandler) UpdatePin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is required", "")
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is not valid", err.Error())
		return
	}

	user := r.Context().Value("user").(*domain.User)

	var req UpdatePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	params := &domain.UpdatePinParams{
		Title:         req.Title,
		Description:   req.Description,
		ImageURL:      req.ImageURL,
		Width:         req.Width,
		Height:        req.Height,
		DominantColor: req.DominantColor,
		Tags:          req.Tags,
	}

	pin, err := h.UC.Update(r.Context(), idInt, user.ID, params)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update pin", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pin)
}

func (h *PinHandler) DeletePin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is required", "")
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "id is not valid", err.Error())
		return
	}

	user := r.Context().Value("user").(*domain.User)

	if err := h.UC.Delete(r.Context(), idInt, user.ID); err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete pin", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
