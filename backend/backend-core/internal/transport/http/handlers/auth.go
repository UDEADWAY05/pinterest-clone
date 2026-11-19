package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pinterest-core/internal/domain"
	"pinterest-core/internal/transport/http/render"
	"pinterest-core/internal/usecase"
	"time"
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
    ID          int64     `json:"id"`
    Email       string    `json:"email"`
    DisplayName string    `json:"display_name"`
    AvatarURL   string    `json:"avatar_url"`
    CreatedAt   time.Time `json:"created_at"`
    Bio         string    `json:"bio"`
}

type authResponse struct {
    User authUserResponse `json:"user"`
}

type authRequest struct {
    Email       string `json:"email"`
    Password    string `json:"password"`
    DisplayName string `json:"display_name"`
    AvatarURL   string `json:"avatar_url"`
    Bio         string `json:"bio"`
}


// @Summary     Register a new user
// @Description Register a new user with the provided credentials
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       user body authRequest true "User registration details"
// @Success     201 {object} authResponse "User registered successfully"
// @Failure     400 {object} errorResponse "Invalid JSON"
// @Failure     500 {object} errorResponse "Internal server error"

// Register регистрирует нового пользователя.
// @Summary      Регистрация
// @Description  Создаёт нового пользователя и возвращает session cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      authRequest  true  "Данные для регистрации"
// @Success      201  {object}  authUserResponse
// @Failure      400  {object}  render.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req authRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", err.Error())
        return
    }

    user, token, err := h.UC.Register(r.Context(), req.Email, req.Password, req.DisplayName, req.AvatarURL, req.Bio)
    if err != nil {
        writeAuthError(w, err)
        return
    }

    setSessionCookie(w, token)
    writeAuthUser(w, user, http.StatusCreated)
}

// Login аутентифицирует пользователя.
// @Summary      Вход
// @Description  Аутентификация по email и паролю, возвращает session cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      authRequest  true  "Email и пароль"
// @Success      200  {object}  authUserResponse
// @Failure      401  {object}  render.ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req authRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        render.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", err.Error())
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

// Logout завершает сессию пользователя.
// @Summary      Выход
// @Description  Удаляет session cookie и завершает сессию
// @Tags         auth
// @Success      204  "No Content"
// @Security     SessionCookie
// @Router       /api/auth/logout [post]
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

// Current возвращает информацию о текущем авторизованном пользователе.
// @Summary      Текущий пользователь
// @Description  Возвращает профиль пользователя по session cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  authUserResponse
// @Failure      401  {object}  render.ErrorResponse
// @Security     SessionCookie
// @Router       /api/user/current [get]
func (h *AuthHandler) Current(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }


    cookie, err := r.Cookie(sessionCookieName)
    if err != nil {
        render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", "")
        return
    }
    fmt.Println("Current user", cookie)

    user, err := h.UC.Current(r.Context(), cookie.Value)
    if err != nil {
        render.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", "")
        return
    }

    fmt.Println(user)

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
        Bio:         user.Bio,
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
    code := "BAD_REQUEST"
    if errors.Is(err, usecase.ErrUnauthorized) {
        status = http.StatusUnauthorized
        code = "UNAUTHORIZED"
    }
    render.Error(w, status, code, err.Error(), "")
}
