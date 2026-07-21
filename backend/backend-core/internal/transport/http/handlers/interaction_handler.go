package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// InteractionHandler — POST /api/interactions.
//
// События приходят от frontend при каждом действии пользователя:
// показ пина (impression), клик, сохранение и т.д.
//
// Эндпоинт не требует auth — события приходят и от анонимов.
// Идентификация идёт по session_id (который frontend хранит в localStorage).
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
	UserID    *int64                 `json:"user_id,omitempty"` // опционально, если пользователь авторизован
	TS        *time.Time             `json:"ts,omitempty"`      // опционально, сервер может проставить свою дату
	Meta      map[string]interface{} `json:"meta,omitempty"`    // произвольные дополнительные данные
}

// ── POST /api/interactions ──────────────────────────────────────────

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

	// Если клиент не прислал ts — ставим время сервера
	ts := time.Now().UTC()
	if req.TS != nil {
		ts = *req.TS
	}

	if err := h.UC.Save(r.Context(), req.UserID, req.SessionID, req.PinID, req.EventType, ts, req.Meta); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	// 202 Accepted — событие принято, но может быть ещё не записано в БД
	// (лежит в буфере и сбросится через пару секунд)
	w.WriteHeader(http.StatusAccepted)
}
