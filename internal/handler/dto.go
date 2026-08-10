package handler

import (
	"time"

	"pdf-box-aws/internal/domain"
)

// FileDTO is the file shape exposed to API clients. It deliberately omits
// S3Key, which is an infrastructure detail that would couple clients to the
// bucket layout, and OwnerID, which is already carried by the JWT.
type FileDTO struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// FileResponse is returned by the single-file endpoints. The presigned URL
// expires, so it is only handed out here and never inside list responses.
type FileResponse struct {
	PresignedURL string   `json:"presigned_url"`
	File         *FileDTO `json:"file"`
}

// ListResponse is returned by the file listing endpoint. NextCursor is empty
// when there are no more pages.
type ListResponse struct {
	Items      []*FileDTO `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

func toFileDTO(file *domain.File) *FileDTO {
	if file == nil {
		return nil
	}
	return &FileDTO{
		ID:        file.ID,
		Filename:  file.Filename,
		Size:      file.Size,
		Status:    string(file.Status),
		CreatedAt: file.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toFileDTOs(files []*domain.File) []*FileDTO {
	dtos := make([]*FileDTO, 0, len(files))
	for _, file := range files {
		dtos = append(dtos, toFileDTO(file))
	}
	return dtos
}
