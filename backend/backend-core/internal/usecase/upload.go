package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"pinterest-core/internal/repository"
)

const maxUploadSize int64 = 10 * 1024 * 1024

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const uploadPathPrefix = "uploads"

type UploadUsecase struct {
	storage  repository.StorageRepository
	cache    repository.CacheRepository
	expiry   time.Duration
	rateRepo *UploadRateRepository
}

type UploadRateRepository struct {
	cache repository.CacheRepository
}

func NewUploadUsecase(
	storage repository.StorageRepository,
	cache repository.CacheRepository,
	expiry time.Duration,
) *UploadUsecase {
	return &UploadUsecase{
		storage:  storage,
		cache:    cache,
		expiry:   expiry,
		rateRepo: &UploadRateRepository{cache: cache},
	}
}

type UploadResult struct {
	UploadURL string `json:"upload_url"` // временный URL для PUT-запроса (15 мин)
	PublicURL string `json:"public_url"` // постоянный URL для чтения (сохраняем в image_url)
	Key       string `json:"key"`        // ключ в S3 (на случай, если клиенту нужен)
	ExpiresIn int    `json:"expires_in"` // секунд до истечения upload_url
}

type UploadRequest struct {
	FileName    string `json:"file_name"`    // оригинальное имя файла (нужно только для расширения)
	ContentType string `json:"content_type"` // MIME-тип (image/jpeg, image/png)
	FileSize    int64  `json:"file_size"`    // размер в байтах
}

func (u *UploadUsecase) PresignUpload(ctx context.Context, userID int64, req UploadRequest) (*UploadResult, error) {
	if req.FileSize <= 0 {
		return nil, errors.New("file_size must be positive")
	}
	if req.FileSize > maxUploadSize {
		return nil, fmt.Errorf("file too large: max %d bytes", maxUploadSize)
	}

	ext, ok := allowedMimeTypes[req.ContentType]
	if !ok {
		return nil, fmt.Errorf("unsupported content type: %q (allowed: jpeg, png, webp)", req.ContentType)
	}

	if err := u.rateRepo.checkRateLimit(ctx, userID); err != nil {
		return nil, err
	}

	uuid, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("generate uuid: %w", err)
	}

	key := fmt.Sprintf("%s/%d/%s%s", uploadPathPrefix, userID, uuid, ext)

	uploadURL, err := u.storage.PresignPutURL(ctx, key, u.expiry)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}

	publicURL := u.storage.PublicURL(key)

	return &UploadResult{
		UploadURL: uploadURL,
		PublicURL: publicURL,
		Key:       key,
		ExpiresIn: int(u.expiry.Seconds()),
	}, nil
}

const maxUploadsPerHour = 50

func (r *UploadRateRepository) checkRateLimit(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("ratelimit:upload:%d", userID)
	ttl := time.Hour

	val, err := r.cache.Increment(ctx, key, ttl)
	if err != nil {
		return nil
	}

	if val > maxUploadsPerHour {
		return errors.New("upload limit exceeded: max 50 uploads per hour")
	}

	return nil
}

func generateUUID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf), nil
}

func ValidateExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedMimeTypes {
		if ext == allowed {
			return true
		}
	}
	return false
}
