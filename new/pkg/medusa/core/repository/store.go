package repository

import (
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"gorm.io/gorm"
)

type Store struct {
	BaseRepo    *Repository
	Transaction TransactionManager
}

func NewStore(db *gorm.DB, logger logger.Logger) *Store {
	return &Store{
		BaseRepo:    NewRepository(db, logger),
		Transaction: NewTransactionManager(db),
	}
}
