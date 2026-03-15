package postgres

import (
	"context"
	"time"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

type directGroupAccessRow struct {
	GroupID      uuid.UUID `gorm:"column:group_id"`
	Role         string    `gorm:"column:role"`
	MemberActive bool      `gorm:"column:member_active"`
}

type inheritedGroupAccessRow struct {
	GroupID        uuid.UUID `gorm:"column:group_id"`
	OrganizationID uuid.UUID `gorm:"column:organization_id"`
}

type channelAdminRow struct {
	ChannelID uuid.UUID `gorm:"column:channel_id"`
}

type guestAccessRow struct {
	InviteID        uuid.UUID  `gorm:"column:invite_id"`
	ChannelID       uuid.UUID  `gorm:"column:channel_id"`
	CanPost         bool       `gorm:"column:can_post"`
	VisibleFromSeq  int64      `gorm:"column:visible_from_seq"`
	InviteExpiresAt time.Time  `gorm:"column:invite_expires_at"`
	InviteRevokedAt *time.Time `gorm:"column:invite_revoked_at"`
	AcceptedAt      *time.Time `gorm:"column:accepted_at"`
	GuestExpiresAt  time.Time  `gorm:"column:guest_expires_at"`
}

func (r *Repo) ResolveGroupAccessForAccount(ctx context.Context, groupID, accountID uuid.UUID) (collabdomain.Access, error) {
	access := collabdomain.Access{GroupID: groupID}
	if groupID == uuid.Nil || accountID == uuid.Nil {
		return access, nil
	}

	var direct directGroupAccessRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.group_account_members AS gam").
		Select("gam.group_id, gam.role, gam.is_active AS member_active").
		Where("gam.group_id = ? AND gam.account_id = ? AND gam.deleted_at IS NULL", groupID, accountID).
		Take(&direct).Error
	if err == nil && direct.MemberActive {
		access.Allowed = true
		access.CanRead = true
		access.CanPost = true
		access.GroupRole = direct.Role
		if direct.Role == "owner" {
			access.CanManage = true
			access.CanModerate = true
		}
	}

	var inherited []inheritedGroupAccessRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.group_organization_members AS gom").
		Select("gom.group_id, gom.organization_id").
		Joins("JOIN iam.memberships AS m ON m.organization_id = gom.organization_id").
		Where("gom.group_id = ? AND gom.deleted_at IS NULL AND gom.is_active = TRUE AND m.account_id = ? AND m.deleted_at IS NULL AND m.is_active = TRUE", groupID, accountID).
		Scan(&inherited).Error; err != nil {
		return access, err
	}
	if len(inherited) > 0 {
		access.Allowed = true
		access.CanRead = true
		access.CanPost = true
		access.OrganizationIDs = make([]uuid.UUID, 0, len(inherited))
		for _, row := range inherited {
			access.OrganizationIDs = append(access.OrganizationIDs, row.OrganizationID)
		}
	}

	return access, nil
}

func (r *Repo) ResolveChannelAccessForAccount(ctx context.Context, channelID, accountID uuid.UUID) (collabdomain.Access, error) {
	access := collabdomain.Access{ChannelID: channelID}
	if channelID == uuid.Nil || accountID == uuid.Nil {
		return access, nil
	}

	channel, err := r.GetChannelByID(ctx, channelID)
	if err != nil || channel == nil {
		return access, err
	}

	access, err = r.ResolveGroupAccessForAccount(ctx, channel.GroupID, accountID)
	if err != nil {
		return collabdomain.Access{}, err
	}
	access.ChannelID = channelID
	if !access.Allowed {
		return access, nil
	}

	// Channel visibility: if channel has organization or account restrictions, user must pass them
	hasOrgs, hasAccounts, err := r.channelHasVisibilityRestrictions(ctx, channelID)
	if err != nil {
		return collabdomain.Access{}, err
	}
	if hasOrgs || hasAccounts {
		allowed, err := r.accountPassesChannelVisibility(ctx, channelID, accountID, access.OrganizationIDs)
		if err != nil {
			return collabdomain.Access{}, err
		}
		if !allowed {
			access.Allowed = false
			return access, nil
		}
	}

	var admin channelAdminRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channel_admins").
		Select("channel_id").
		Where("channel_id = ? AND account_id = ?", channelID, accountID).
		Take(&admin).Error; err == nil {
		access.ChannelAdmin = true
		access.CanModerate = true
		access.CanManage = true
	}

	return access, nil
}

func (r *Repo) channelHasVisibilityRestrictions(ctx context.Context, channelID uuid.UUID) (hasOrgs, hasAccounts bool, err error) {
	var orgCount int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channel_organizations").
		Where("channel_id = ?", channelID).
		Count(&orgCount).Error; err != nil {
		return false, false, err
	}
	var accCount int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channel_accounts").
		Where("channel_id = ?", channelID).
		Count(&accCount).Error; err != nil {
		return false, false, err
	}
	return orgCount > 0, accCount > 0, nil
}

func (r *Repo) accountPassesChannelVisibility(ctx context.Context, channelID, accountID uuid.UUID, userOrgIDs []uuid.UUID) (bool, error) {
	var directCount int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channel_accounts").
		Where("channel_id = ? AND account_id = ?", channelID, accountID).
		Count(&directCount).Error; err != nil {
		return false, err
	}
	if directCount > 0 {
		return true, nil
	}
	if len(userOrgIDs) == 0 {
		return false, nil
	}
	var orgMatchCount int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channel_organizations").
		Where("channel_id = ? AND organization_id IN ?", channelID, userOrgIDs).
		Count(&orgMatchCount).Error; err != nil {
		return false, err
	}
	return orgMatchCount > 0, nil
}

func (r *Repo) ResolveChannelAccessForGuest(ctx context.Context, channelID, guestID uuid.UUID) (collabdomain.Access, error) {
	access := collabdomain.Access{ChannelID: channelID, IsGuest: true}
	if channelID == uuid.Nil || guestID == uuid.Nil {
		return access, nil
	}

	var row guestAccessRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("auth.guest_identities AS g").
		Select("gi.id AS invite_id, g.channel_id, gi.can_post, gi.visible_from_seq, gi.expires_at AS invite_expires_at, gi.revoked_at AS invite_revoked_at, gi.accepted_at, g.expires_at AS guest_expires_at").
		Joins("JOIN collab.guest_invites AS gi ON gi.id = g.invite_id").
		Where("g.id = ? AND g.channel_id = ?", guestID, channelID).
		Take(&row).Error; err != nil {
		return access, nil
	}

	now := time.Now().UTC()
	if row.InviteRevokedAt != nil || row.AcceptedAt == nil || !row.InviteExpiresAt.After(now) || !row.GuestExpiresAt.After(now) {
		return access, nil
	}

	access.Allowed = true
	access.CanRead = true
	access.CanPost = row.CanPost
	access.VisibleFromSeq = row.VisibleFromSeq
	access.InviteID = row.InviteID
	return access, nil
}
