package postgres

import (
	"context"
	"fmt"
	"time"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/mapper"
	"github.com/google/uuid"
)

func (r *SessionRepo) Create(ctx context.Context, session *authdomain.RefreshSession) error {
	m := mapper.ToDBRefreshSessionForCreate(session)
	if m == nil {
		return autherrors.InvalidInput("Session is required")
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Create(m).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %v", autherrors.ErrConflict, err)
		}
		return err
	}
	return nil
}

func (r *SessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.RefreshSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

func (r *SessionRepo) ReplaceToken(ctx context.Context, sessionID uuid.UUID, newTokenHash string) error {
	now := time.Now()
	return r.dbFrom(ctx).WithContext(ctx).
		Model(&dbmodel.RefreshSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{
			"token_hash": newTokenHash,
			"updated_at": now,
		}).Error
}
