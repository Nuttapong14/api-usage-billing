package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Uploader uploads binary payloads to storage.
type Uploader interface {
	Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
}

// MinioConfig holds MinIO connection configuration.
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

// MinioUploader uploads files to MinIO.
type MinioUploader struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// NewMinioUploader creates a new MinIO uploader.
func NewMinioUploader(cfg MinioConfig) (*MinioUploader, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, err
		}
	}

	return &MinioUploader{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
		useSSL:   cfg.UseSSL,
	}, nil
}

// Upload uploads a payload and returns a presigned URL.
func (u *MinioUploader) Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	_, err := u.client.PutObject(ctx, u.bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	url, err := u.client.PresignedGetObject(ctx, u.bucket, objectName, 24*time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (u *MinioUploader) endpointURL() string {
	scheme := "http"
	if u.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, u.endpoint)
}
