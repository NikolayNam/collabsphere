package postgres

import (
	"context"
	"errors"
	"time"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type guestInviteRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	ChannelID      uuid.UUID  `gorm:"column:channel_id"`
	Email          string     `gorm:"column:email"`
	CanPost        bool       `gorm:"column:can_post"`
	VisibleFromSeq int64      `gorm:"column:visible_from_seq"`
	ExpiresAt      time.Time  `gorm:"column:expires_at"`
	AcceptedAt     *time.Time `gorm:"column:accepted_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	InvitedBy      uuid.UUID  `gorm:"column:invited_by"`
}

type guestIdentityRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	InviteID    uuid.UUID `gorm:"column:invite_id"`
	ChannelID   uuid.UUID `gorm:"column:channel_id"`
	Email       string    `gorm:"column:email"`
	DisplayName string    `gorm:"column:display_name"`
	ExpiresAt   time.Time `gorm:"column:expires_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (r *Repo) CreateGuestInvite(ctx context.Context, invite collabdomain.GuestInvite, tokenHash string) (*collabdomain.GuestInvite, error) {
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now().UTC()
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.guest_invites").Create(map[string]any{
		"id":               invite.ID,
		"channel_id":       invite.ChannelID,
		"email":            invite.Email,
		"token_hash":       tokenHash,
		"can_post":         invite.CanPost,
		"visible_from_seq": invite.VisibleFromSeq,
		"expires_at":       invite.ExpiresAt,
		"created_at":       invite.CreatedAt,
		"invited_by":       invite.InvitedBy,
	}).Error; err != nil {
		return nil, err
	}
	return r.GetGuestInviteByID(ctx, invite.ID)
}

func (r *Repo) GetGuestInviteByID(ctx context.Context, inviteID uuid.UUID) (*collabdomain.GuestInvite, error) {
	var row guestInviteRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.guest_invites").
		Select("id, channel_id, email, can_post, visible_from_seq, expires_at, accepted_at, revoked_at, created_at, invited_by").
		Where("id = ?", inviteID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapGuestInvite(row), nil
}

func (r *Repo) GetGuestInviteByTokenHash(ctx context.Context, tokenHash string) (*collabdomain.GuestInvite, error) {
	var row guestInviteRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.guest_invites").
		Select("id, channel_id, email, can_post, visible_from_seq, expires_at, accepted_at, revoked_at, created_at, invited_by").
		Where("token_hash = ?", tokenHash).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapGuestInvite(row), nil
}

func (r *Repo) ExchangeGuestInvite(ctx context.Context, tokenHash, displayName, sessionTokenHash string, userAgent, ip *string, guestExpiresAt, sessionExpiresAt, now time.Time) (*collabdomain.GuestInvite, *collabdomain.GuestIdentity, uuid.UUID, error) {
	var invite *collabdomain.GuestInvite
	var identity *collabdomain.GuestIdentity
	sessionID := uuid.Nil

	err := r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inviteRow guestInviteRow
		if err := tx.Table("collab.guest_invites").
			Select("id, channel_id, email, can_post, visible_from_seq, expires_at, accepted_at, revoked_at, created_at, invited_by").
			Where("token_hash = ?", tokenHash).
			Take(&inviteRow).Error; err != nil {
			return err
		}
		invite = mapGuestInvite(inviteRow)

		var identityRow guestIdentityRow
		err := tx.Table("auth.guest_identities").
			Select("id, invite_id, channel_id, email, display_name, expires_at, created_at").
			Where("invite_id = ?", invite.ID).
			Take(&identityRow).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			identity = &collabdomain.GuestIdentity{
				ID:          uuid.New(),
				InviteID:    invite.ID,
				ChannelID:   invite.ChannelID,
				Email:       invite.Email,
				DisplayName: displayName,
				ExpiresAt:   guestExpiresAt,
				CreatedAt:   now,
			}
			if err := tx.Table("auth.guest_identities").Create(map[string]any{
				"id":           identity.ID,
				"invite_id":    identity.InviteID,
				"channel_id":   identity.ChannelID,
				"email":        identity.Email,
				"display_name": identity.DisplayName,
				"accepted_at":  now,
				"expires_at":   identity.ExpiresAt,
				"created_at":   identity.CreatedAt,
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}
			if err := tx.Table("collab.guest_invites").Where("id = ?", invite.ID).Updates(map[string]any{
				"accepted_at":          now,
				"accepted_by_guest_id": identity.ID,
				"updated_at":           now,
			}).Error; err != nil {
				return err
			}
		} else {
			identity = &collabdomain.GuestIdentity{
				ID:          identityRow.ID,
				InviteID:    identityRow.InviteID,
				ChannelID:   identityRow.ChannelID,
				Email:       identityRow.Email,
				DisplayName: identityRow.DisplayName,
				ExpiresAt:   identityRow.ExpiresAt,
				CreatedAt:   identityRow.CreatedAt,
			}
			if err := tx.Table("auth.guest_identities").Where("id = ?", identity.ID).Updates(map[string]any{
				"last_seen_at": now,
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}
		}

		sessionID = uuid.New()
		return tx.Table("auth.guest_sessions").Create(map[string]any{
			"id":         sessionID,
			"guest_id":   identity.ID,
			"token_hash": sessionTokenHash,
			"user_agent": userAgent,
			"ip_address": ip,
			"expires_at": sessionExpiresAt,
			"created_at": now,
			"updated_at": now,
		}).Error
	})
	if err != nil {
		return nil, nil, uuid.Nil, err
	}
	return invite, identity, sessionID, nil
}

func mapGuestInvite(row guestInviteRow) *collabdomain.GuestInvite {
	return &collabdomain.GuestInvite{
		ID:             row.ID,
		ChannelID:      row.ChannelID,
		Email:          row.Email,
		CanPost:        row.CanPost,
		VisibleFromSeq: row.VisibleFromSeq,
		ExpiresAt:      row.ExpiresAt,
		AcceptedAt:     row.AcceptedAt,
		RevokedAt:      row.RevokedAt,
		CreatedAt:      row.CreatedAt,
		InvitedBy:      row.InvitedBy,
	}
}

func (r *Repo) GetGuestIdentityByID(ctx context.Context, guestID uuid.UUID) (*collabdomain.GuestIdentity, error) {
	var row guestIdentityRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("auth.guest_identities").
		Select("id, invite_id, channel_id, email, display_name, expires_at, created_at").
		Where("id = ?", guestID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &collabdomain.GuestIdentity{
		ID:          row.ID,
		InviteID:    row.InviteID,
		ChannelID:   row.ChannelID,
		Email:       row.Email,
		DisplayName: row.DisplayName,
		ExpiresAt:   row.ExpiresAt,
		CreatedAt:   row.CreatedAt,
	}, nil
}
