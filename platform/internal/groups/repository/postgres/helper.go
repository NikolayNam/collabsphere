package postgres

import (
	"context"

	"gorm.io/gorm"

	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
)

func (r *GroupRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := dbtx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}
