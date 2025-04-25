package store

import (
	"github.com/imlargo/go-api/internal/store/repositories"
	"gorm.io/gorm"
)

type Storage struct {
	Users repositories.UserRepository
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{
		Users: repositories.NewUserRepository(db),
	}
}
