package repository

import (
	"pdf-box-aws/internal/domain"
	"time"
)

type fileItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	S3Key     string `dynamodbav:"s3Key"`
	Filename  string `dynamodbav:"filename"`
	Size      int64  `dynamodbav:"size"`
	Status    string `dynamodbav:"status"`
	CreatedAt int64  `dynamodbav:"createdAt"`
	ExpiresAt int64  `dynamodbav:"expiresAt,omitempty"`
}

func (f *fileItem) ToDomainFile() *domain.File {
	return &domain.File{
		ID:        f.SK[len("FILE#"):], // Remove the "FILE#" prefix
		OwnerID:   f.PK[len("USER#"):], // Remove the "USER#" prefix
		S3Key:     f.S3Key,
		Filename:  f.Filename,
		Size:      f.Size,
		Status:    domain.Status(f.Status),
		CreatedAt: time.Unix(f.CreatedAt, 0),
	}
}

func FromDomainFile(f *domain.File) *fileItem {
	return &fileItem{
		PK:        "USER#" + f.OwnerID,
		SK:        "FILE#" + f.ID,
		S3Key:     f.S3Key,
		Filename:  f.Filename,
		Size:      f.Size,
		Status:    string(f.Status),
		CreatedAt: f.CreatedAt.Unix(),
	}
}
