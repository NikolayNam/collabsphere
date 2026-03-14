package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	apperrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxInvitationListRows = 500

func (r *MembershipRepo) CreateInvitation(ctx context.Context, invitation *memberdomain.OrganizationInvitation) error {
	if invitation == nil {
		return apperrors.InvalidInput("Organization invitation is required")
	}
	model := &dbmodel.OrganizationInvitation{
		ID:                  invitation.ID(),
		OrganizationID:      invitation.OrganizationID().UUID(),
		Email:               invitation.Email().String(),
		Role:                string(invitation.Role()),
		TokenHash:           invitation.TokenHash(),
		InviterAccountID:    invitation.InviterAccountID(),
		AcceptedByAccountID: invitation.AcceptedByAccountID(),
		AcceptedAt:          invitation.AcceptedAt(),
		RevokedByAccountID:  invitation.RevokedByAccountID(),
		RevokedAt:           invitation.RevokedAt(),
		ExpiresAt:           invitation.ExpiresAt(),
		CreatedAt:           invitation.CreatedAt(),
		UpdatedAt:           derefTime(invitation.UpdatedAt(), invitation.CreatedAt()),
	}
	err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.InvitationAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return apperrors.InvalidInput("Organization or inviter account not found")
		}
		return fmt.Errorf("create invitation: %w", err)
	}
	return nil
}

func (r *MembershipRepo) SaveInvitation(ctx context.Context, invitation *memberdomain.OrganizationInvitation) error {
	if invitation == nil {
		return apperrors.InvalidInput("Organization invitation is required")
	}
	updates := map[string]any{
		"accepted_by_account_id": invitation.AcceptedByAccountID(),
		"accepted_at":            invitation.AcceptedAt(),
		"revoked_by_account_id":  invitation.RevokedByAccountID(),
		"revoked_at":             invitation.RevokedAt(),
		"updated_at":             derefTime(invitation.UpdatedAt(), invitation.CreatedAt()),
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.organization_invitations").
		Where("id = ?", invitation.ID()).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("save invitation: %w", err)
	}
	return nil
}

func (r *MembershipRepo) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*memberdomain.OrganizationInvitation, error) {
	var model dbmodel.OrganizationInvitation
	err := r.dbFrom(ctx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ?", tokenHash).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapInvitationModel(model)
}

func (r *MembershipRepo) ListInvitations(ctx context.Context, orgID orgdomain.OrganizationID) ([]memberdomain.OrganizationInvitation, error) {
	var rows []dbmodel.OrganizationInvitation
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.organization_invitations").
		Where("organization_id = ?", orgID.UUID()).
		Order("created_at desc").
		Limit(maxInvitationListRows).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]memberdomain.OrganizationInvitation, 0, len(rows))
	for _, row := range rows {
		invitation, err := mapInvitationModel(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *invitation)
	}
	return out, nil
}

func (r *MembershipRepo) RevokeExpiredPendingInvitations(ctx context.Context, orgID orgdomain.OrganizationID, email accdomain.Email, actorAccountID uuid.UUID, now time.Time) error {
	updates := map[string]any{
		"revoked_by_account_id": actorAccountID,
		"revoked_at":            now,
		"updated_at":            now,
	}
	return r.dbFrom(ctx).WithContext(ctx).
		Table("iam.organization_invitations").
		Where("organization_id = ? AND email = ? AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at <= ?", orgID.UUID(), email.String(), now).
		Updates(updates).Error
}

func mapInvitationModel(model dbmodel.OrganizationInvitation) (*memberdomain.OrganizationInvitation, error) {
	orgID, err := orgdomain.OrganizationIDFromUUID(model.OrganizationID)
	if err != nil {
		return nil, err
	}
	email, err := accdomain.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}
	return memberdomain.RehydrateOrganizationInvitation(memberdomain.RehydrateOrganizationInvitationParams{
		ID:                  model.ID,
		OrganizationID:      orgID,
		Email:               email,
		Role:                memberdomain.ParseMembershipRole(model.Role),
		TokenHash:           model.TokenHash,
		InviterAccountID:    model.InviterAccountID,
		AcceptedByAccountID: cloneUUID(model.AcceptedByAccountID),
		AcceptedAt:          cloneTime(model.AcceptedAt),
		RevokedByAccountID:  cloneUUID(model.RevokedByAccountID),
		RevokedAt:           cloneTime(model.RevokedAt),
		ExpiresAt:           model.ExpiresAt,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           cloneTime(&model.UpdatedAt),
	})
}

func cloneUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}
