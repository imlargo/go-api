package repository

import (
	"context"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"gorm.io/gorm"
)

type ctxKey string

const txKey ctxKey = "tx"

type Repository struct {
	entity any
	logger *logger.Logger
	db     *gorm.DB
}

func NewRepository(db *gorm.DB, logger *logger.Logger, opts ...RepositoryOption) *Repository {
	r := &Repository{
		db:     db,
		logger: logger,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// DB devuelve la conexión apropiada (tx si existe en contexto, o db normal)
func (r *Repository) DB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}
