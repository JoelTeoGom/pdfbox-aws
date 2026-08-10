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

// presignTTL is how long a generated URL stays usable.
const presignTTL = 15 * time.Minute

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

// DeleteFile removes an object. S3 answers with success when the key does not
// exist, so callers can retry this safely.
func (s *S3Service) DeleteFile(ctx context.Context, s3Key string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	return err
}

// PresignDownload returns a short lived URL that fetches an object.
func (s *S3Service) PresignDownload(ctx context.Context, s3Key string) (string, error) {
	if s3Key == "" {
		return "", errors.New("s3Key cannot be empty")
	}

	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(presignTTL))
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	return req.URL, nil
}

// PresignUpload returns a short lived URL that stores an object. Content-Type
// and Content-Length take part in the signature, so a client cannot upload a
// different content type or a different number of bytes than the ones declared:
// S3 rejects the request with a 403 before storing anything, which is what keeps
// an oversized body from ever reaching the bucket.
func (s *S3Service) PresignUpload(ctx context.Context, s3Key, contentType string, contentLength int64) (string, error) {
	if s3Key == "" {
		return "", errors.New("s3Key cannot be empty")
	}
	if contentType == "" {
		return "", errors.New("contentType cannot be empty")
	}
	if contentLength <= 0 {
		return "", errors.New("contentLength must be positive")
	}

	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(s3Key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}, s3.WithPresignExpires(presignTTL))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}

	return req.URL, nil
}
