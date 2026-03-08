package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
)

func (r *AccountRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}
