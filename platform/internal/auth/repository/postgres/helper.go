package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
)

func (r *SessionRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func (r *SessionRepo) withinTransaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return fn(gormTx.WithContext(ctx))
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx.WithContext(ctx))
	})
}
