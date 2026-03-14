package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *SessionRepo) Create(ctx context.Context, session *authdomain.RefreshSession) error {
	m := mapper.ToDBRefreshSessionForCreate(session)
	if m == nil {
		return autherrors.InvalidInput("Session is required")
	}

	tokenModel := &dbmodel.RefreshSessionToken{
		ID:        uuid.New(),
		SessionID: session.ID(),
		TokenHash: session.TokenHash(),
		CreatedAt: session.CreatedAt(),
	}

	return r.withinTransaction(ctx, func(db *gorm.DB) error {
		if err := db.Create(m).Error; err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: %v", autherrors.ErrConflict, err)
			}
			return err
		}
		if err := db.Create(tokenModel).Error; err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: %v", autherrors.ErrConflict, err)
			}
			return err
		}
		return nil
	})
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

func (r *SessionRepo) RotateByRefreshToken(ctx context.Context, presentedTokenHash, newTokenHash string, now time.Time) (*authdomain.RefreshSession, error) {
	var result *authdomain.RefreshSession

	err := r.withinTransaction(ctx, func(db *gorm.DB) error {
		var token dbmodel.RefreshSessionToken
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ?", presentedTokenHash).
			Take(&token).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		var session dbmodel.RefreshSession
		err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", token.SessionID).
			Take(&session).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
			return nil
		}

		if token.UsedAt != nil || session.TokenHash != presentedTokenHash {
			return revokeSession(db, session.ID, now)
		}

		if err := db.Model(&dbmodel.RefreshSessionToken{}).
			Where("id = ? AND used_at IS NULL", token.ID).
			Updates(map[string]any{"used_at": now}).Error; err != nil {
			return err
		}

		if err := db.Create(&dbmodel.RefreshSessionToken{
			ID:        uuid.New(),
			SessionID: session.ID,
			TokenHash: newTokenHash,
			CreatedAt: now,
		}).Error; err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: %v", autherrors.ErrConflict, err)
			}
			return err
		}

		update := db.Model(&dbmodel.RefreshSession{}).
			Where("id = ? AND token_hash = ? AND revoked_at IS NULL", session.ID, presentedTokenHash).
			Updates(map[string]any{
				"token_hash": newTokenHash,
				"updated_at": now,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected != 1 {
			return revokeSession(db, session.ID, now)
		}

		session.TokenHash = newTokenHash
		updatedAt := now
		session.UpdatedAt = &updatedAt

		mapped, err := mapper.ToDomainRefreshSession(&session)
		if err != nil {
			return err
		}
		result = mapped
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func revokeSession(db *gorm.DB, sessionID uuid.UUID, now time.Time) error {
	return db.Model(&dbmodel.RefreshSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]any{
			"revoked_at": now,
			"updated_at": now,
		}).Error
}
