package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OneTimeCodeRepo struct {
	db *gorm.DB
}

func NewOneTimeCodeRepo(db *gorm.DB) *OneTimeCodeRepo {
	return &OneTimeCodeRepo{db: db}
}

func (r *OneTimeCodeRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func (r *OneTimeCodeRepo) Create(ctx context.Context, record *ports.OneTimeCodeRecord) error {
	if record == nil {
		return autherrors.InvalidInput("One-time code is required")
	}
	model := &dbmodel.OneTimeCode{
		ID:           record.ID,
		Purpose:      record.Purpose,
		CodeHash:     record.CodeHash,
		AccountID:    record.AccountID,
		Provider:     record.Provider,
		Intent:       record.Intent,
		IsNewAccount: record.IsNewAccount,
		ExpiresAt:    record.ExpiresAt,
		UsedAt:       record.UsedAt,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %w", autherrors.ErrConflict, err)
		}
		return err
	}
	return nil
}

func (r *OneTimeCodeRepo) GetByCodeHash(ctx context.Context, purpose, codeHash string) (*ports.OneTimeCodeRecord, error) {
	var model dbmodel.OneTimeCode
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("purpose = ? AND code_hash = ?", purpose, codeHash).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ports.OneTimeCodeRecord{
		ID:           model.ID,
		Purpose:      model.Purpose,
		CodeHash:     model.CodeHash,
		AccountID:    model.AccountID,
		Provider:     model.Provider,
		Intent:       model.Intent,
		IsNewAccount: model.IsNewAccount,
		ExpiresAt:    model.ExpiresAt,
		UsedAt:       model.UsedAt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

func (r *OneTimeCodeRepo) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	result := r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.OneTimeCode{}).
		Where("id = ? AND used_at IS NULL", id).
		Updates(map[string]any{"used_at": at, "updated_at": at})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}
