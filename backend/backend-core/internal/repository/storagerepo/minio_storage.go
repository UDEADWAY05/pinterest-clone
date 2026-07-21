package storagerepo

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewMinioStorage(client *minio.Client, bucket, publicBaseURL string) *MinioStorage {
	return &MinioStorage{
		client:        client,
		bucket:        bucket,
		publicBaseURL: publicBaseURL,
	}
}

func (s *MinioStorage) PresignPutURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.Presign(ctx, "PUT", s.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign put (bucket=%q, key=%q): %w", s.bucket, key, err)
	}

	return presignedURL.String(), nil
}

func (s *MinioStorage) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, s.bucket, key)
}
