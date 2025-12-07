package store

import (
	"github.com/imlargo/go-api/internal/repository"
	medusarepo "github.com/imlargo/go-api/pkg/medusa/core/repository"
)

type Store struct {
	medusarepo.Store
	UserRepository repository.UserRepository
}

func NewStore(store medusarepo.Store) *Store {
	return &Store{
		Store:          store,
		UserRepository: repository.NewUserRepository(*store.BaseRepo),
	}
}
