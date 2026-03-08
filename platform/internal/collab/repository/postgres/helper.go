package postgres

import (
	"context"

	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"gorm.io/gorm"
)

func (r *Repo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := dbtx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}
