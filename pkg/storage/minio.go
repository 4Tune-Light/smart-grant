package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStorage interface {
	Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error)
}

type MinioStorage struct {
	client *minio.Client
	bucket string
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

func NewMinio(cfg Config) (*MinioStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &MinioStorage{client: client, bucket: cfg.Bucket}, nil
}

func (m *MinioStorage) Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload to minio: %w", err)
	}

	url := fmt.Sprintf("/api/v1/files/%s", objectPath)
	return url, nil
}
