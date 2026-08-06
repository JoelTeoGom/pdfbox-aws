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
}

func (api *API) Get(ctx context.Context, userID, fileID string) (*domain.File, error) {
	return api.repo.Get(ctx, userID, fileID)
}

func (api *API) Save(ctx context.Context, f *domain.File) error {
	return api.repo.Save(ctx, f)
}

func (api *API) List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error) {
	return api.repo.List(ctx, userID, cursor, limit)
}

func (api *API) MarkDeleted(ctx context.Context, userID, fileID string) error {
	if err := api.repo.MarkDeleted(ctx, userID, fileID); err != nil {
		return err
	}
	return api.s3.DeleteFile(ctx, fileID)
}
