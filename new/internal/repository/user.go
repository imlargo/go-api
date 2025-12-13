package repository

import (
	"context"

	"github.com/imlargo/go-api/internal/models"
	medusarepo "github.com/imlargo/go-api/pkg/medusa/core/repository"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*models.User, error)
}

type userRepository struct {
	*medusarepo.Repository
}

func NewUserRepository(repo *medusarepo.Repository) UserRepository {
	return &userRepository{Repository: repo}
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.DB(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
