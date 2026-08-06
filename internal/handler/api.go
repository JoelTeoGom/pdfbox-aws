package handler

import (
	"context"
	"pdf-box-aws/internal/domain"
)

type API struct {
	repo FileRepository
}

func NewAPI(repo FileRepository) *API {
	return &API{
		repo: repo,
	}
}

type FileRepository interface {
	Get(ctx context.Context, userID, fileID string) (*domain.File, error)
	Save(ctx context.Context, f *domain.File) error
	List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error)
	MarkDeleted(ctx context.Context, userID, fileID string) error
}
