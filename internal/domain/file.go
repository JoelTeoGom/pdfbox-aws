package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusUploaded Status = "uploaded"
	StatusRejected Status = "rejected"
	StatusDeleted  Status = "deleted"
	StatusTrashed  Status = "trashed"
)

const (
	// MaxFileSize is the largest upload accepted, in bytes (50 MB)
	MaxFileSize int64 = 50 * 1024 * 1024

	// MimeTypePDF is the only content type this service stores
	MimeTypePDF = "application/pdf"

	DefaultTrashRetention = 24 * time.Hour
)

type File struct {
	ID        string
	OwnerID   string
	S3Key     string
	Filename  string
	Size      int64
	Mime      string
	Status    Status
	CreatedAt time.Time

	TrashedAt      time.Time
	TrashExpiresAt time.Time

	S3Deleted bool
}

// NewFileID returns an identifier that sorts chronologically as a string. The
// listing pages straight off the sort key, so the ID is what carries the
// creation order: a random UUID here would page in arbitrary order.
func NewFileID() string {
	return ulid.Make().String()
}

func (f *File) CanBeDownloaded() bool {
	return f.Status == StatusUploaded
}
func (f *File) CanBeTrashed() bool {
	return f.Status == StatusUploaded
}

func (f *File) CanBeRestored() bool {
	return f.Status == StatusTrashed
}

func (f *File) IsTrashed() bool {
	return f.Status == StatusTrashed
}

// S3KeyFor builds the object key under which a user's file is stored. This is
// the single definition of the bucket layout; ParseS3Key is its inverse, and
// the two must be changed together.
func S3KeyFor(ownerID, fileID string) string {
	return fmt.Sprintf("users/%s/%s.pdf", ownerID, fileID)
}

// ParseS3Key extracts the owner and file IDs from an object key. Keys arriving
// from S3 event notifications are URL-encoded, so they are decoded first.
func ParseS3Key(s3Key string) (ownerID, fileID string, err error) {
	decodedKey, err := url.QueryUnescape(s3Key)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(decodedKey, "/")
	if len(parts) != 3 || parts[0] != "users" || !strings.HasSuffix(parts[2], ".pdf") {
		return "", "", fmt.Errorf("unexpected key format: %s", s3Key)
	}

	return parts[1], strings.TrimSuffix(parts[2], ".pdf"), nil
}
