package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
)

// UploadHandler — обработчик POST /api/uploads/presign.
type UploadHandler struct {
	UC     *usecase.UploadUsecase
	AuthUC *usecase.AuthUsecase
}

func NewUploadHandler(uc *usecase.UploadUsecase, authUC *usecase.AuthUsecase) *UploadHandler {
	return &UploadHandler{UC: uc, AuthUC: authUC}
}

// PresignUpload генерирует presigned URL для загрузки изображения.
// @Summary      Получить presigned URL
// @Description  Генерирует временный URL для прямой загрузки файла в S3/MinIO. Файл не проходит через сервер.
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        body  body      usecase.UploadRequest  true  "Информация о файле"
// @Success      200  {object}  usecase.UploadResult
// @Failure      400  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/uploads/presign [post]
func (h *UploadHandler) PresignUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUploadUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	var req usecase.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	result, err := h.UC.PresignUpload(r.Context(), userID, req)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	render.JSON(w, http.StatusOK, result)
}

func (h *UploadHandler) getUploadUserID(r *http.Request) (int64, error) {
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
