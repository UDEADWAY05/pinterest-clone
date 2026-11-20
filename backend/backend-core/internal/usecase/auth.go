package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"pinterest-core/internal/domain"
	"pinterest-core/internal/repository"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthUsecase struct {
	Users     repository.UserRepository
	Sessions  repository.SessionRepository
	SessionTTL time.Duration
}

func NewAuthUsecase(users repository.UserRepository, sessions repository.SessionRepository, sessionTTL time.Duration) *AuthUsecase {
	return &AuthUsecase{
		Users:      users,
		Sessions:   sessions,
		SessionTTL: sessionTTL,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email, password, displayName, avatarURL, bio string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)
	avatarURL = strings.TrimSpace(avatarURL)
	bio = strings.TrimSpace(bio)

	if email == "" || password == "" || displayName == "" {
		return nil, "", errors.New("email, password and display_name are required")
	}

	if _, err := u.Users.GetUserByEmail(ctx, email); err == nil {
		return nil, "", errors.New("email is already registered")
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		AvatarURL:    avatarURL,
		Bio:          bio,
	}

	if err := u.Users.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := u.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, "", errors.New("email and password are required")
	}

	user, err := u.Users.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrUnauthorized
	}

	token, err := u.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (u *AuthUsecase) Current(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}

	userID, err := u.Sessions.GetUserIDByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, ErrUnauthorized
	}

	user, err := u.Users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return user, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return u.Sessions.DeleteSessionByTokenHash(ctx, hashToken(token))
}

func (u *AuthUsecase) createSession(ctx context.Context, userID int64) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	if err := u.Sessions.CreateSession(ctx, hashToken(token), userID, time.Now().Add(u.sessionTTL())); err != nil {
		return "", err
	}

	return token, nil
}

func (u *AuthUsecase) sessionTTL() time.Duration {
	if u.SessionTTL <= 0 {
		return 7 * 24 * time.Hour
	}
	return u.SessionTTL
}

func generateSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
