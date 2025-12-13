package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// TransactionManager maneja transacciones de manera centralizada
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithTransactionOpts(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.WithTransactionOpts(ctx, nil, fn)
}

func (tm *transactionManager) WithTransactionOpts(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	// Si ya hay una transacción en el contexto, reutilizarla (transacciones anidadas)
	if _, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return fn(ctx)
	}

	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	}, opts)
}
