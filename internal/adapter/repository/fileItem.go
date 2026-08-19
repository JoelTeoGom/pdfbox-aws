package repository

import (
	"pdf-box-aws/internal/domain"
	"time"
)

const sweepPrefix = "SWEEP#"

type fileItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	S3Key     string `dynamodbav:"s3Key"`
	Filename  string `dynamodbav:"filename"`
	Size      int64  `dynamodbav:"size"`
	Mime      string `dynamodbav:"mime"`
	Status    string `dynamodbav:"status"`
	CreatedAt int64  `dynamodbav:"createdAt"`

	TrashedAt      int64 `dynamodbav:"trashedAt,omitempty"`
	TrashExpiresAt int64 `dynamodbav:"trashExpiresAt,omitempty"`
	S3Deleted      bool  `dynamodbav:"s3Deleted,omitempty"`
	ExpiresAt      int64 `dynamodbav:"expiresAt,omitempty"`

	SweepPK string `dynamodbav:"gsiPK,omitempty"`
	SweepSK int64  `dynamodbav:"gsiSK,omitempty"`
}

func (f *fileItem) ToDomainFile() *domain.File {
	return &domain.File{
		ID:             f.SK[len("FILE#"):], // Remove the "FILE#" prefix
		OwnerID:        f.PK[len("USER#"):], // Remove the "USER#" prefix
		S3Key:          f.S3Key,
		Filename:       f.Filename,
		Size:           f.Size,
		Mime:           f.Mime,
		Status:         domain.Status(f.Status),
		CreatedAt:      timeFromEpoch(f.CreatedAt),
		TrashedAt:      timeFromEpoch(f.TrashedAt),
		TrashExpiresAt: timeFromEpoch(f.TrashExpiresAt),
		S3Deleted:      f.S3Deleted,
	}
}

func FromDomainFile(f *domain.File) *fileItem {
	sweepPartitionKey, sweepDueAt := sweepKeysFor(f)

	return &fileItem{
		PK:             "USER#" + f.OwnerID,
		SK:             "FILE#" + f.ID,
		S3Key:          f.S3Key,
		Filename:       f.Filename,
		Size:           f.Size,
		Mime:           f.Mime,
		Status:         string(f.Status),
		CreatedAt:      epochFromTime(f.CreatedAt),
		TrashedAt:      epochFromTime(f.TrashedAt),
		TrashExpiresAt: epochFromTime(f.TrashExpiresAt),
		S3Deleted:      f.S3Deleted,
		SweepPK:        sweepPartitionKey,
		SweepSK:        sweepDueAt,
	}
}

func sweepKeysFor(f *domain.File) (partitionKey string, dueAt int64) {
	if f.Status == domain.StatusUploaded {
		return "", 0
	}
	if f.Status == domain.StatusDeleted && f.S3Deleted {
		return "", 0
	}

	dueAt = epochFromTime(f.TrashExpiresAt)
	if dueAt == 0 {
		dueAt = epochFromTime(f.CreatedAt)
	}

	return sweepPrefix + string(f.Status), dueAt
}

func epochFromTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func timeFromEpoch(seconds int64) time.Time {
	if seconds == 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
}
