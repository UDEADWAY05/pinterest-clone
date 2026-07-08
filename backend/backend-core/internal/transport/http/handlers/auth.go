package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/usecase"
)

const sessionCookieName = "session"
const sessionCookieMaxAge = 7 * 24 * 60 * 60

type AuthHandler struct {
	UC *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{UC: uc}
}

type authUserResponse struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	CreatedAt   time.Time  `json:"created_at"`
}

type authResponse struct {
	User authUserResponse `json:"user"`
}

type authRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.UC.Register(r.Context(), req.Email, req.Password, req.DisplayName, req.AvatarURL)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	setSessionCookie(w, token)
	writeAuthUser(w, user, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, token, err := h.UC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	setSessionCookie(w, token)
	writeAuthUser(w, user, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.UC.Logout(r.Context(), cookie.Value)
	}

	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Current(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.UC.Current(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	writeAuthUser(w, user, http.StatusOK)
}

func writeAuthUser(w http.ResponseWriter, user *domain.User, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(authUserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		CreatedAt:   user.CreatedAt,
	})
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   sessionCookieMaxAge,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, usecase.ErrUnauthorized) {
		status = http.StatusUnauthorized
	}
	http.Error(w, err.Error(), status)
}
