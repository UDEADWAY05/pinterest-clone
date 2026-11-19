package handlers

import (
	"net/http"
	"strconv"

	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// RecommendationHandler — GET /api/recommendations + GET /api/pins/{id}/related.
type RecommendationHandler struct {
	UC     *usecase.RecommendationUsecase
	AuthUC *usecase.AuthUsecase
}

func NewRecommendationHandler(uc *usecase.RecommendationUsecase, authUC *usecase.AuthUsecase) *RecommendationHandler {
	return &RecommendationHandler{UC: uc, AuthUC: authUC}
}

// GetFeed возвращает персонализированную ленту.
// @Summary      Рекомендации
// @Description  Возвращает персонализированную ленту пинов с пагинацией. Если пользователь авторизован — лента строится на основе его взаимодействий.
// @Tags         recommendations
// @Produce      json
// @Param        limit   query  int  false  "Количество пинов (default 20)"
// @Param        offset  query  int  false  "Смещение (default 0)"
// @Success      200  {object}  render.JSONResponse{data=[]domain.Pin}
// @Router       /api/recommendations [get]
func (h *RecommendationHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	limit := getQueryInt(r, "limit", 20)
	offset := getQueryInt(r, "offset", 0)
	sessionID := r.URL.Query().Get("session_id")

	var userID int64
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		user, err := h.AuthUC.Current(r.Context(), cookie.Value)
		if err == nil {
			userID = user.ID
		}
		if sessionID == "" {
			sessionID = cookie.Value
		}
	}

	pins, err := h.UC.GetFeed(r.Context(), userID, sessionID, limit, offset)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load recommendations", err.Error())
		return
	}

	pins, err = h.UC.EnrichPins(r.Context(), pins, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load recommendations", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

// GetRelated возвращает похожие на указанный пин.
// @Summary      Похожие пины
// @Description  Возвращает пины, похожие на указанный (по embedding или тегам)
// @Tags         recommendations
// @Produce      json
// @Param        id     path  int  true  "ID пина"
// @Param        limit  query  int  false  "Количество (default 10)"
// @Success      200  {object}  render.JSONResponse{data=[]domain.Pin}
// @Router       /api/pins/{id}/related [get]
func (h *RecommendationHandler) GetRelated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pinID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pin id", "")
		return
	}

	limit := getQueryInt(r, "limit", 10)

	pins, err := h.UC.GetRelated(r.Context(), pinID, limit)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load related pins", err.Error())
		return
	}

	pins, err = h.UC.EnrichPins(r.Context(), pins, currentViewerID(r, h.AuthUC))
	if err != nil {
		render.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load related pins", err.Error())
		return
	}

	render.JSON(w, http.StatusOK, pins)
}

func getQueryInt(r *http.Request, name string, def int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 {
		return def
	}
	return n
}
