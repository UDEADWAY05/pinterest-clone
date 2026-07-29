package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// InteractionHandler — POST /api/interactions.
type InteractionHandler struct {
	UC *usecase.InteractionUsecase
}

func NewInteractionHandler(uc *usecase.InteractionUsecase) *InteractionHandler {
	return &InteractionHandler{UC: uc}
}

type interactionRequest struct {
	EventType string                 `json:"event_type"`
	PinID     int64                  `json:"pin_id"`
	SessionID string                 `json:"session_id"`
	UserID    *int64                 `json:"user_id,omitempty"`
	TS        *time.Time             `json:"ts,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// Save сохраняет событие поведения пользователя.
// @Summary      Отправить событие
// @Description  Сохраняет событие поведения пользователя (показ, клик, сохранение и т.д.). Событие попадает в буфер и сбрасывается в БД пачкой.
// @Tags         interactions
// @Accept       json
// @Success      202  "Accepted — событие принято в буфер"
// @Failure      400  {object}  render.ErrorResponse
// @Router       /api/interactions [post]
func (h *InteractionHandler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req interactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	ts := time.Now().UTC()
	if req.TS != nil {
		ts = *req.TS
	}

	if err := h.UC.Save(r.Context(), req.UserID, req.SessionID, req.PinID, req.EventType, ts, req.Meta); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
