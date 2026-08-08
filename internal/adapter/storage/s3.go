package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewS3Service(ctx context.Context, bucket string) *S3Service {
	cfg, _ := config.LoadDefaultConfig(ctx)

	s3Client := s3.NewFromConfig(cfg)
	presignCli := s3.NewPresignClient(s3Client)
	return &S3Service{
		s3Client:      s3Client,
		presignClient: presignCli,
		bucket:        bucket,
	}
}

func (s *S3Service) DeleteFile(ctx context.Context, fileID string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileID),
	})
	return err
}

func (s *S3Service) GeneratePresignedURL(ctx context.Context, method, fileID string) (string, error) {
	if fileID == "" {
		return "", errors.New("fileID cannot be empty")
	}
	if method != "GET" && method != "PUT" {
		return "", errors.New("unsupported method")
	}

	if method == "GET" {
		req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(fileID),
		}, s3.WithPresignExpires(15*time.Minute))
		if err != nil {
			return "", fmt.Errorf("presign get: %w", err)
		}
		return req.URL, nil
	}

	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileID),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}

	return req.URL, nil
}
