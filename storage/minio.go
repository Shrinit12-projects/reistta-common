// minio.go contains application logic.
package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Provider         string
	Endpoint         string
	AccessKey        string
	SecretKey        string
	UseSSL           bool
	Bucket           string
	Region           string
	AutoCreateBucket bool
}

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

func NewMinio(ctx context.Context, cfg MinioConfig) (*MinioClient, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("minio config is incomplete")
	}

	// Initialize MinIO client with static credentials.
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio: %w", err)
	}

	// Ensure bucket exists at startup to avoid runtime failures.
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if !cfg.AutoCreateBucket {
			return nil, fmt.Errorf("bucket %q does not exist", cfg.Bucket)
		}
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &MinioClient{Client: client, Bucket: cfg.Bucket}, nil
}

func (m *MinioClient) PresignPut(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	url, err := m.Client.PresignedPutObject(ctx, m.Bucket, objectKey, expiry)
	if err != nil {
		return "", err
	}
	if len(reqParams) > 0 {
		url.RawQuery = reqParams.Encode()
	}
	return url.String(), nil
}

func (m *MinioClient) PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	url, err := m.Client.PresignedGetObject(ctx, m.Bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
