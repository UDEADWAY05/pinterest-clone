package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// BoardHandler — HTTP-обработчики для досок и save/unsave пинов.
type BoardHandler struct {
	UC     *usecase.BoardUsecase
	AuthUC *usecase.AuthUsecase
}

func NewBoardHandler(uc *usecase.BoardUsecase, authUC *usecase.AuthUsecase) *BoardHandler {
	return &BoardHandler{UC: uc, AuthUC: authUC}
}

// ── DTO (Request/Response структуры) ─────────────────────────────────

type createBoardRequest struct {
	Title     string `json:"title"`
	IsPrivate bool   `json:"is_private"`
}

// ── POST /api/boards — создать доску ─────────────────────────────────

func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	var req createBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	board, err := h.UC.Create(r.Context(), userID, req.Title, req.IsPrivate)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	render.JSON(w, http.StatusCreated, board)
}

// ── GET /api/boards/me — мои доски ───────────────────────────────────

func (h *BoardHandler) ListMyBoards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	boards, err := h.UC.ListUserBoards(r.Context(), userID)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load boards", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, boards)
}

// ── POST /api/boards/{boardId}/pins/{pinId} — сохранить пин ──────────

func (h *BoardHandler) SavePinToBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	boardID, err := strconv.ParseInt(r.PathValue("boardId"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid board id", "")
		return
	}

	pinID, err := strconv.ParseInt(r.PathValue("pinId"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pin id", "")
		return
	}

	bp, err := h.UC.SavePinToBoard(r.Context(), userID, boardID, pinID)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	render.JSON(w, http.StatusCreated, bp)
}

// ── DELETE /api/boards/{boardId}/pins/{pinId} — удалить пин из доски ─

func (h *BoardHandler) RemovePinFromBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	boardID, err := strconv.ParseInt(r.PathValue("boardId"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid board id", "")
		return
	}

	pinID, err := strconv.ParseInt(r.PathValue("pinId"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pin id", "")
		return
	}

	if err := h.UC.RemovePinFromBoard(r.Context(), userID, boardID, pinID); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── POST/DELETE /api/boards/{boardId}/pins/{pinId} ─────────────────
//
// Комбинированный handler: POST = сохранить, DELETE = удалить.
// Это нужно потому, что http.ServeMux не умеет различать методы на одном пути.
// В Go 1.22+ есть mux.HandleFunc("POST /path", ...), но для совместимости
// с существующим стилем кода используем этот подход.

func (h *BoardHandler) HandlePinInBoard(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.SavePinToBoard(w, r)
	case http.MethodDelete:
		h.RemovePinFromBoard(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ── GET /api/boards/{boardId}/pins — пины в доске ────────────────────

func (h *BoardHandler) GetBoardPins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	boardID, err := strconv.ParseInt(r.PathValue("boardId"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid board id", "")
		return
	}

	pins, err := h.UC.GetBoardPins(r.Context(), boardID)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

// ── Helper — достаёт userID из сессии ─────────────────────────────────
//
// В существующем коде authUC.Current может использовать cookie "session".
// Этот хелпер унифицирует получение текущего пользователя.
func (h *BoardHandler) getCurrentUserID(r *http.Request) (int64, error) {
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
