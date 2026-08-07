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

type createBoardRequest struct {
	Title     string `json:"title"`
	IsPrivate bool   `json:"is_private"`
}

// CreateBoard создаёт новую доску.
// @Summary      Создать доску
// @Description  Создаёт новую доску для текущего пользователя
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        body  body      createBoardRequest  true  "Данные доски"
// @Success      201  {object}  render.JSONResponse{data=domain.Board}
// @Failure      400  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/boards [post]
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

// ListMyBoards возвращает доски текущего пользователя.
// @Summary      Мои доски
// @Description  Возвращает все доски текущего авторизованного пользователя
// @Tags         boards
// @Produce      json
// @Success      200  {object}  render.JSONResponse{data=[]domain.Board}
// @Security     SessionCookie
// @Router       /api/boards/me [get]
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

// SavePinToBoard сохраняет пин в доску.
// @Summary      Сохранить пин
// @Description  Сохраняет пин в указанную доску (только для владельца доски)
// @Tags         boards
// @Produce      json
// @Success      201  {object}  render.JSONResponse{data=domain.BoardPin}
// @Failure      400  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/boards/{boardId}/pins/{pinId} [post]
func (h *BoardHandler) SavePinToBoard(w http.ResponseWriter, r *http.Request) {
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

// RemovePinFromBoard удаляет пин из доски.
// @Summary      Удалить пин из доски
// @Description  Удаляет пин из указанной доски (только для владельца доски)
// @Tags         boards
// @Produce      json
// @Success      204  "No Content"
// @Failure      400  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/boards/{boardId}/pins/{pinId} [delete]
func (h *BoardHandler) RemovePinFromBoard(w http.ResponseWriter, r *http.Request) {
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

// GetBoardPins возвращает пины в доске.
// @Summary      Пины доски
// @Description  Возвращает все пины, сохранённые в указанной доске
// @Tags         boards
// @Produce      json
// @Param        boardId  path  int  true  "ID доски"
// @Success      200  {object}  render.JSONResponse{data=[]domain.Pin}
// @Router       /api/boards/{boardId}/pins [get]
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

	pins, err = h.UC.EnrichPins(r.Context(), pins, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load pins", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

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
