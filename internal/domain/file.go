package domain

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusReady    Status = "ready"
	StatusRejected Status = "rejected"
	StatusDeleted  Status = "deleted"
)

type File struct {
	ID        string
	OwnerID   string
	S3Key     string
	Filename  string
	Size      int64
	Status    Status
	CreatedAt time.Time
}

func (f *File) CanBeDownloaded() bool {
	return f.Status == StatusReady
}
