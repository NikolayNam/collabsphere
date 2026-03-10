package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"gorm.io/gorm"
)

func (r *Repo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}
