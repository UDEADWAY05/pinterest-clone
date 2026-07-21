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

// ── POST /api/uploads/presign ────────────────────────────────────────
//
// Клиент присылает информацию о будущем файле, backend возвращает
// presigned URL, по которому клиент загружает файл напрямую в S3.
//
// Request:
//
//	{
//	  "file_name": "my_photo.jpg",
//	  "content_type": "image/jpeg",
//	  "file_size": 5242880
//	}
//
// Response:
//
//	{
//	  "upload_url": "https://...X-Amz-Signature=...",
//	  "public_url": "http://localhost:9000/pinterest-images/uploads/1/abc...jpg",
//	  "key": "uploads/1/abc...jpg",
//	  "expires_in": 900
//	}
func (h *UploadHandler) PresignUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// ── Получаем пользователя из сессии ───────────────────────────
	userID, err := h.getUploadUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), "")
		return
	}

	// ── Декодируем запрос ─────────────────────────────────────────
	var req usecase.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON", err.Error())
		return
	}

	// ── Вызываем usecase ───────────────────────────────────────────
	result, err := h.UC.PresignUpload(r.Context(), userID, req)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "")
		return
	}

	// ── Успех ─────────────────────────────────────────────────────
	render.JSON(w, http.StatusOK, result)
}

// getUploadUserID — helper, извлекает userID из session cookie.
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
