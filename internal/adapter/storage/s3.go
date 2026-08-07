package storage

import (
	"context"
	"strings"
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
		Bucket: aws.String("your-bucket-name"),
		Key:    aws.String(fileID),
	})
	return err
}

func (s *S3Service) GeneratePresignedURL(ctx context.Context, fileID string) (string, error) {
	req, _ := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("amzn-s3-demo-bucket"),
		Key:    aws.String("myKey"),
		Body:   strings.NewReader("EXPECTED CONTENTS"),
	})
	str, err := req.Presign(15 * time.Minute)
	return req.URL, nil
}
