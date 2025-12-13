package service

import (
	"context"

	"github.com/imlargo/go-api/internal/models"
)

type UserService interface {
	GetUserByID(userID uint) (*models.User, error)
}

type userService struct {
	*Service
}

func NewUserService(container *Service) UserService {
	return &userService{
		Service: container,
	}
}

func (s *userService) GetUserByID(userID uint) (*models.User, error) {

	user, err := s.store.UserRepository.GetByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
