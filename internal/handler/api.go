package handler

import (
	"context"
	"pdf-box-aws/internal/domain"
	"time"

	"github.com/google/uuid"
)

type API struct {
	repo FileRepository
	s3   S3service
}

func NewAPI(repo FileRepository, s3 S3service) *API {
	return &API{
		repo: repo,
		s3:   s3,
	}
}

type FileRepository interface {
	Get(ctx context.Context, userID, fileID string) (*domain.File, error)
	Save(ctx context.Context, f *domain.File) error
	List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error)
	UpdateStatus(ctx context.Context, userID, fileID string, status domain.Status) error
}
type TokenValidator interface {
	ValidateToken(token string) (*domain.Claims, error)
}

type S3service interface {
	DeleteFile(ctx context.Context, s3Key string) error
	PresignDownload(ctx context.Context, s3Key string) (string, error)
	PresignUpload(ctx context.Context, s3Key, contentType string, contentLength int64) (string, error)
}

// Get returns the stored file along with a presigned URL to download it.
func (api *API) Get(ctx context.Context, userID, fileID string) (*domain.File, string, error) {
	if fileID == "" {
		return nil, "", domain.ErrFileNotFound
	}
	if userID == "" {
		return nil, "", domain.ErrUnauthorized
	}

	file, err := api.repo.Get(ctx, userID, fileID)
	if err != nil {
		return nil, "", err
	}
	if file == nil {
		return nil, "", domain.ErrFileNotFound
	}
	presignedURL, err := api.s3.PresignDownload(ctx, file.S3Key)
	if err != nil {
		return nil, "", err
	}

	return file, presignedURL, nil
}

// Save records the pending file and returns it along with a presigned URL the
// client uses to upload the bytes directly to S3.
func (api *API) Save(ctx context.Context, userId, filename string, size int64, mime string) (*domain.File, string, error) {
	if userId == "" {
		return nil, "", domain.ErrUnauthorized
	}
	if filename == "" || size <= 0 || mime == "" {
		return nil, "", domain.ErrInvalidInput
	}

	if size > domain.MaxFileSize {
		return nil, "", domain.ErrFileTooLarge
	}

	if mime != domain.MimeTypePDF {
		return nil, "", domain.ErrInvalidMimeType
	}
	fileID := uuid.New().String()
	file := &domain.File{
		ID:        fileID,
		OwnerID:   userId,
		S3Key:     domain.S3KeyFor(userId, fileID),
		Filename:  filename,
		Size:      size,
		Mime:      mime,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}
	if err := api.repo.Save(ctx, file); err != nil {
		return nil, "", err
	}
	presignedURL, err := api.s3.PresignUpload(ctx, file.S3Key, file.Mime, file.Size)
	if err != nil {
		return nil, "", err
	}
	return file, presignedURL, nil
}

func (api *API) List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error) {
	if userID == "" {
		return nil, "", domain.ErrUnauthorized
	}

	return api.repo.List(ctx, userID, cursor, limit)
}

func (api *API) MarkDeleted(ctx context.Context, userID, fileID string) error {
	if userID == "" {
		return domain.ErrUnauthorized
	}
	if fileID == "" {
		return domain.ErrFileNotFound
	}
	if err := api.repo.UpdateStatus(ctx, userID, fileID, domain.StatusDeleted); err != nil {
		return err
	}
	// Dynamo is updated first so the API stops serving the file straight away.
	// A failure here only leaks the bytes, which the reconciliation job sweeps.
	return api.s3.DeleteFile(ctx, domain.S3KeyFor(userID, fileID))
}
