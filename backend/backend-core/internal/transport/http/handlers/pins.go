package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pinterest-core/internal/usecase"
	"strconv"
)

type PinHandler struct {
	UC *usecase.PinUsecase
	AuthUC *usecase.AuthUsecase
}

func NewPinHandler(uc *usecase.PinUsecase, authuc *usecase.AuthUsecase) *PinHandler {
	return &PinHandler{UC: uc, AuthUC: authuc}
}

func (h *PinHandler) ListPins(w http.ResponseWriter, r *http.Request) {
	pins, err := h.UC.List(r.Context())
	if err != nil {
		 http.Error(w, "failed to load pins", http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pins)
}

func (h *PinHandler) GetPinByID(w http.ResponseWriter, r *http.Request) {
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

type CreatePinRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	DominantColor string `json:"dominant_color"`
}

func (h *PinHandler) CreatePin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")

	if err != nil {
		http.Error(w, "unauthorized: missing session token", http.StatusUnauthorized) // 401
		return
	}

	var req CreatePinRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.AuthUC.Current(r.Context(), cookie.Value)

	if err != nil {
		http.Error(w, "unauthorized: invalid or expired session", http.StatusUnauthorized) // 401
		return
	}

	authorID := user.ID

	pin, err := h.UC.Create(r.Context(), req.Title, req.Description, req.DominantColor, req.ImageURL, req.Width, req.Height, authorID)

	if err != nil {
		fmt.Print(err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pin)
}
