package service

import (
    "context"
	"app/internal/model"
	"app/internal/repository"
)

type ContentService interface {
	GetContent(ctx context.Context, id int64) (*model.Content, error)
}
func NewContentService(
    service *Service,
    contentRepository repository.ContentRepository,
) ContentService {
	return &contentService{
		Service:        service,
		contentRepository: contentRepository,
	}
}

type contentService struct {
	*Service
	contentRepository repository.ContentRepository
}

func (s *contentService) GetContent(ctx context.Context, id int64) (*model.Content, error) {
	return s.contentRepository.GetContent(ctx, id)
}
