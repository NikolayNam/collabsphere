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

type ExternalIdentityRepo struct {
	db *gorm.DB
}

func NewExternalIdentityRepo(db *gorm.DB) *ExternalIdentityRepo {
	return &ExternalIdentityRepo{db: db}
}

func (r *ExternalIdentityRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func (r *ExternalIdentityRepo) GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*ports.ExternalIdentityRecord, error) {
	var model dbmodel.ExternalIdentity
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("provider = ? AND external_subject = ?", provider, externalSubject).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapExternalIdentity(&model), nil
}

func (r *ExternalIdentityRepo) Create(ctx context.Context, record *ports.ExternalIdentityRecord) error {
	if record == nil {
		return autherrors.InvalidInput("External identity is required")
	}
	model := &dbmodel.ExternalIdentity{
		ID:              record.ID,
		Provider:        record.Provider,
		ExternalSubject: record.ExternalSubject,
		AccountID:       record.AccountID,
		Email:           record.Email,
		EmailVerified:   record.EmailVerified,
		DisplayName:     record.DisplayName,
		ClaimsJSON:      record.ClaimsJSON,
		LastLoginAt:     record.LastLoginAt,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %w", autherrors.ErrConflict, err)
		}
		return err
	}
	return nil
}

func (r *ExternalIdentityRepo) TouchLogin(ctx context.Context, id uuid.UUID, email *string, emailVerified bool, displayName *string, claimsJSON string, at time.Time) error {
	updates := map[string]any{
		"email":          email,
		"email_verified": emailVerified,
		"display_name":   displayName,
		"claims_json":    claimsJSON,
		"last_login_at":  at,
		"updated_at":     at,
	}
	return r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.ExternalIdentity{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func mapExternalIdentity(model *dbmodel.ExternalIdentity) *ports.ExternalIdentityRecord {
	if model == nil {
		return nil
	}
	return &ports.ExternalIdentityRecord{
		ID:              model.ID,
		Provider:        model.Provider,
		ExternalSubject: model.ExternalSubject,
		AccountID:       model.AccountID,
		Email:           model.Email,
		EmailVerified:   model.EmailVerified,
		DisplayName:     model.DisplayName,
		ClaimsJSON:      model.ClaimsJSON,
		LastLoginAt:     model.LastLoginAt,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
