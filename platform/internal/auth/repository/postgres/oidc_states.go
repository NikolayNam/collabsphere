package postgres

import (
	"context"
	"errors"
	"time"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OIDCStateRepo struct {
	db *gorm.DB
}

func NewOIDCStateRepo(db *gorm.DB) *OIDCStateRepo {
	return &OIDCStateRepo{db: db}
}

func (r *OIDCStateRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func (r *OIDCStateRepo) CreateState(ctx context.Context, record *ports.OAuthStateRecord) error {
	if record == nil {
		return autherrors.InvalidInput("OAuth state is required")
	}
	model := &dbmodel.OAuthState{
		ID:        record.ID,
		Provider:  record.Provider,
		StateHash: record.StateHash,
		ExpiresAt: record.ExpiresAt,
		UsedAt:    record.UsedAt,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	return r.dbFrom(ctx).WithContext(ctx).Create(model).Error
}

func (r *OIDCStateRepo) CreateNonce(ctx context.Context, record *ports.OIDCNonceRecord) error {
	if record == nil {
		return autherrors.InvalidInput("OIDC nonce is required")
	}
	model := &dbmodel.OIDCNonce{
		ID:           record.ID,
		Provider:     record.Provider,
		OAuthStateID: record.OAuthStateID,
		NonceHash:    record.NonceHash,
		ExpiresAt:    record.ExpiresAt,
		UsedAt:       record.UsedAt,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
	return r.dbFrom(ctx).WithContext(ctx).Create(model).Error
}

func (r *OIDCStateRepo) GetStateByHash(ctx context.Context, provider, stateHash string) (*ports.OAuthStateRecord, error) {
	var model dbmodel.OAuthState
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("provider = ? AND state_hash = ?", provider, stateHash).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ports.OAuthStateRecord{
		ID:        model.ID,
		Provider:  model.Provider,
		StateHash: model.StateHash,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *OIDCStateRepo) GetNonceByStateID(ctx context.Context, provider string, stateID uuid.UUID) (*ports.OIDCNonceRecord, error) {
	var model dbmodel.OIDCNonce
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("provider = ? AND oauth_state_id = ?", provider, stateID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ports.OIDCNonceRecord{
		ID:           model.ID,
		Provider:     model.Provider,
		OAuthStateID: model.OAuthStateID,
		NonceHash:    model.NonceHash,
		ExpiresAt:    model.ExpiresAt,
		UsedAt:       model.UsedAt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

func (r *OIDCStateRepo) MarkStateUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.OAuthState{}).
		Where("id = ?", id).
		Updates(map[string]any{"used_at": at, "updated_at": at}).Error
}

func (r *OIDCStateRepo) MarkNonceUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.OIDCNonce{}).
		Where("id = ?", id).
		Updates(map[string]any{"used_at": at, "updated_at": at}).Error
}
