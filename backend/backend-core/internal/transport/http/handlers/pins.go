package handlers

import (
	"encoding/json"
	"net/http"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
	"strconv"
)

type PinHandler struct {
	UC     *usecase.PinUsecase
	AuthUC *usecase.AuthUsecase
}

func NewPinHandler(uc *usecase.PinUsecase, authuc *usecase.AuthUsecase) *PinHandler {
	return &PinHandler{UC: uc, AuthUC: authuc}
}

func (h *PinHandler) ListPins(w http.ResponseWriter, r *http.Request) {
	pins, err := h.UC.List(r.Context())
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", err.Error())
		return
	}

	// Используем render.JSON вместо ручного кодирования
	render.JSON(w, http.StatusOK, pins)
}

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
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get pin", err.Error())
		return
	}

	// Данные будут красиво обернуты в {"data": {...}}
	render.JSON(w, http.StatusOK, pin)
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

func (h *PinHandler) CreatePin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing session token", err.Error())
		return
	}

	var req CreatePinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	user, err := h.AuthUC.Current(r.Context(), cookie.Value)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired session", err.Error())
		return
	}

	authorID := user.ID

	pin, err := h.UC.Create(r.Context(), req.Title, req.Description, req.DominantColor, req.ImageURL, req.Width, req.Height, authorID, req.Tags)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid data", err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, pin)
}
