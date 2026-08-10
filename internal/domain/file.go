package domain

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusUploaded Status = "uploaded"
	StatusRejected Status = "rejected"
	StatusDeleted  Status = "deleted"
)

const (
	// MaxFileSize is the largest upload accepted, in bytes (50 MB)
	MaxFileSize int64 = 50 * 1024 * 1024

	// MimeTypePDF is the only content type this service stores
	MimeTypePDF = "application/pdf"
)

type FileResponse struct {
	PresignedURL string `json:"presigned_url"`
	FileData     *File  `json:"file_data"`
}

type File struct {
	ID        string
	OwnerID   string
	S3Key     string
	Filename  string
	Size      int64
	Mime      string
	Status    Status
	CreatedAt time.Time
}

func (f *File) CanBeDownloaded() bool {
	return f.Status == StatusUploaded
}
