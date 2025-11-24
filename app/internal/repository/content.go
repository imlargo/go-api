package repository

import (
    "context"
	"app/internal/model"
)

type ContentRepository interface {
	GetContent(ctx context.Context, id int64) (*model.Content, error)
}

func NewContentRepository(
	repository *Repository,
) ContentRepository {
	return &contentRepository{
		Repository: repository,
	}
}

type contentRepository struct {
	*Repository
}

func (r *contentRepository) GetContent(ctx context.Context, id int64) (*model.Content, error) {
	var content model.Content

	return &content, nil
}
