package tx

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

func TxFromContext(ctx context.Context) *gorm.DB {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return nil
	}
	t, _ := v.(*gorm.DB)
	return t
}

type Manager struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}
