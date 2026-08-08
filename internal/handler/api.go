package handler

import (
	"context"
	"pdf-box-aws/internal/domain"
)

type API struct {
	repo FileRepository
	auth TokenValidator
	s3   S3service
}

func NewAPI(repo FileRepository, auth TokenValidator, s3 S3service) *API {
	return &API{
		repo: repo,
		auth: auth,
		s3:   s3,
	}
}

type FileRepository interface {
	Get(ctx context.Context, userID, fileID string) (*domain.File, error)
	Save(ctx context.Context, f *domain.File) error
	List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error)
	MarkDeleted(ctx context.Context, userID, fileID string) error
}
type TokenValidator interface {
	ValidateToken(token string) (*domain.Claims, error)
}

type S3service interface {
	DeleteFile(ctx context.Context, fileID string) error
	GeneratePresignedURL(ctx context.Context, method, fileID string) (string, error)
}

func (api *API) Get(ctx context.Context, userID, fileID string) (*domain.FileResponse, error) {
	if fileID == "" {
		return nil, domain.ErrFileNotFound
	}
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	file, err := api.repo.Get(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, domain.ErrFileNotFound
	}
	presignedURL, err := api.s3.GeneratePresignedURL(ctx, "GET", fileID)
	if err != nil {
		return nil, err
	}

	return &domain.FileResponse{
		PresignedURL: presignedURL,
		FileData:     file,
	}, nil
}

func (api *API) Save(ctx context.Context, userId, filename string) (*domain.FileResponse, error) {
	if userId == "" {
		return nil, domain.ErrUnauthorized
	}
	file := &domain.File{
		OwnerID:  userId,
		Filename: filename,
		Status:   domain.StatusPending,
	}
	if err := api.repo.Save(ctx, file); err != nil {
		return nil, err
	}
	presignedURL, err := api.s3.GeneratePresignedURL(ctx, "PUT", file.ID)
	if err != nil {
		return nil, err
	}
	return &domain.FileResponse{
		PresignedURL: presignedURL,
		FileData:     file,
	}, nil
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
	if err := api.repo.MarkDeleted(ctx, userID, fileID); err != nil {
		return err
	}
	return api.s3.DeleteFile(ctx, fileID)
}
