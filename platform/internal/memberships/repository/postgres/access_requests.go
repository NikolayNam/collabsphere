package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxAccessRequestListRows = 500

func (r *MembershipRepo) CreateAccessRequest(ctx context.Context, req *memberdomain.OrganizationAccessRequest) error {
	if req == nil {
		return apperrors.InvalidInput("Organization access request is required")
	}
	model := &dbmodel.OrganizationAccessRequest{
		ID:                 req.ID(),
		OrganizationID:     req.OrganizationID().UUID(),
		RequesterAccountID: req.RequesterAccountID(),
		RequestedRole:      string(req.RequestedRole()),
		Message:            req.Message(),
		Status:             string(req.Status()),
		ReviewerAccountID:  req.ReviewerAccountID(),
		ReviewNote:         req.ReviewNote(),
		ReviewedAt:         req.ReviewedAt(),
		CreatedAt:          req.CreatedAt(),
		UpdatedAt:          derefTime(req.UpdatedAt(), req.CreatedAt()),
	}
	err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.AccessRequestAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return apperrors.InvalidInput("Organization or account not found")
		}
		return fmt.Errorf("create access request: %w", err)
	}
	return nil
}

func (r *MembershipRepo) SaveAccessRequest(ctx context.Context, req *memberdomain.OrganizationAccessRequest) error {
	if req == nil {
		return apperrors.InvalidInput("Organization access request is required")
	}
	updates := map[string]any{
		"status":              string(req.Status()),
		"reviewer_account_id": req.ReviewerAccountID(),
		"review_note":         req.ReviewNote(),
		"reviewed_at":         req.ReviewedAt(),
		"updated_at":          derefTime(req.UpdatedAt(), req.CreatedAt()),
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.organization_access_requests").
		Where("id = ? AND organization_id = ?", req.ID(), req.OrganizationID().UUID()).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("save access request: %w", err)
	}
	return nil
}

func (r *MembershipRepo) GetAccessRequestByID(ctx context.Context, orgID orgdomain.OrganizationID, requestID uuid.UUID) (*memberdomain.OrganizationAccessRequest, error) {
	var model dbmodel.OrganizationAccessRequest
	err := r.dbFrom(ctx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("organization_id = ? AND id = ?", orgID.UUID(), requestID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapAccessRequestModel(model)
}

func (r *MembershipRepo) ListAccessRequests(ctx context.Context, orgID orgdomain.OrganizationID) ([]memberdomain.OrganizationAccessRequest, error) {
	var rows []dbmodel.OrganizationAccessRequest
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.organization_access_requests").
		Where("organization_id = ?", orgID.UUID()).
		Order("created_at desc").
		Limit(maxAccessRequestListRows).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]memberdomain.OrganizationAccessRequest, 0, len(rows))
	for _, row := range rows {
		req, err := mapAccessRequestModel(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *req)
	}
	return out, nil
}

func mapAccessRequestModel(model dbmodel.OrganizationAccessRequest) (*memberdomain.OrganizationAccessRequest, error) {
	orgID, err := orgdomain.OrganizationIDFromUUID(model.OrganizationID)
	if err != nil {
		return nil, err
	}
	return memberdomain.RehydrateOrganizationAccessRequest(memberdomain.RehydrateOrganizationAccessRequestParams{
		ID:               model.ID,
		OrganizationID:   orgID,
		RequesterAccount: model.RequesterAccountID,
		RequestedRole:    memberdomain.ParseMembershipRole(model.RequestedRole),
		Message:          model.Message,
		Status:           memberdomain.AccessRequestStatus(model.Status),
		ReviewerAccount:  cloneUUID(model.ReviewerAccountID),
		ReviewNote:       model.ReviewNote,
		ReviewedAt:       cloneTime(model.ReviewedAt),
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        cloneTime(&model.UpdatedAt),
	})
}
