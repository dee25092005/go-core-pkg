package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	localconfig "todo-backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService interface {
	UploadFile(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
}

type s3Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewR2Storage(ctx context.Context, cfg localconfig.Config) (StorageService, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID),
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2.AccessKeyID,
				cfg.R2.AccessKeySecret,
				"",
			)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &s3Storage{
		client:     client,
		bucketName: cfg.R2.BucketName,
		publicURL:  strings.TrimRight(cfg.R2.PublicURL, "/"),
	}, nil
}

func (s *s3Storage) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	fileUrl := fmt.Sprintf("%s/%s", s.publicURL, key)
	return fileUrl, nil
}

func (s *s3Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
