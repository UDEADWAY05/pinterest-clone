package storage

import (
	"context"
	"fmt"

	"pinterest-core/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3 struct {
	Client *minio.Client
}


func New(ctx context.Context, cfg *config.Config) (*S3, error) {
	if cfg.S3Endpoint == "" || cfg.S3Bucket == "" {
		return nil, fmt.Errorf("S3_ENDPOINT and S3_BUCKET must be set for storage")
	}

	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	bucketExists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}

	if !bucketExists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{
			Region: cfg.S3Region,
		}); err != nil {
			return nil, fmt.Errorf("create bucket %q: %w", cfg.S3Bucket, err)
		}
	}

	return &S3{Client: client}, nil
}

func (s *S3) Close(ctx context.Context) error {
	return nil
}
