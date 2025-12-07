package repository

import (
	"context"

	medusarepo "github.com/imlargo/go-api/pkg/medusa/core/repository"
)

type User struct {
	ID    uint
	Name  string
	Email string
}

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
}

type userRepository struct {
	medusarepo.Repository
}

func NewUserRepository(repo medusarepo.Repository) UserRepository {
	return &userRepository{Repository: repo}
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := r.DB(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
