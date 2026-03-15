package postgres

import (
	"context"
	"errors"
	"strings"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationRoleRepo struct {
	db *gorm.DB
}

func NewOrganizationRoleRepo(db *gorm.DB) *OrganizationRoleRepo {
	return &OrganizationRoleRepo{db: db}
}

func (r *OrganizationRoleRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func (r *OrganizationRoleRepo) Create(ctx context.Context, role *memberDomain.OrganizationRole) error {
	if role == nil {
		return memberDomain.ErrOrganizationRoleInvalid
	}
	model := toDBOrganizationRole(role)
	if err := r.dbFrom(ctx).WithContext(ctx).Table("org.organization_roles").Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return memberDomain.ErrOrganizationRoleCodeInvalid
		}
		return err
	}
	return nil
}

func (r *OrganizationRoleRepo) Save(ctx context.Context, role *memberDomain.OrganizationRole) error {
	if role == nil {
		return memberDomain.ErrOrganizationRoleInvalid
	}
	model := toDBOrganizationRole(role)
	result := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_roles").
		Where("organization_id = ? AND id = ?", model.OrganizationID, model.ID).
		Updates(map[string]any{
			"name":        model.Name,
			"description": model.Description,
			"base_role":   model.BaseRole,
			"updated_at":  model.UpdatedAt,
			"deleted_at":  model.DeletedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *OrganizationRoleRepo) GetByID(ctx context.Context, orgID orgDomain.OrganizationID, roleID uuid.UUID) (*memberDomain.OrganizationRole, error) {
	var model dbmodel.OrganizationRole
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_roles").
		Where("organization_id = ? AND id = ?", orgID.UUID(), roleID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainOrganizationRole(&model)
}

func (r *OrganizationRoleRepo) GetByCode(ctx context.Context, orgID orgDomain.OrganizationID, code string) (*memberDomain.OrganizationRole, error) {
	var model dbmodel.OrganizationRole
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_roles").
		Where("organization_id = ? AND code = ? AND deleted_at IS NULL", orgID.UUID(), code).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainOrganizationRole(&model)
}

func (r *OrganizationRoleRepo) List(ctx context.Context, orgID orgDomain.OrganizationID, includeDeleted bool) ([]*memberDomain.OrganizationRole, error) {
	query := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_roles").
		Where("organization_id = ?", orgID.UUID()).
		Order("created_at ASC, id ASC")
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}
	var models []dbmodel.OrganizationRole
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*memberDomain.OrganizationRole, 0, len(models))
	for i := range models {
		role, err := toDomainOrganizationRole(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, nil
}

func (r *OrganizationRoleRepo) CountMembersWithRole(ctx context.Context, orgID orgDomain.OrganizationID, roleCode string) (int64, error) {
	var count int64
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Where("organization_id = ? AND role = ? AND deleted_at IS NULL", orgID.UUID(), strings.TrimSpace(roleCode)).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func toDBOrganizationRole(r *memberDomain.OrganizationRole) *dbmodel.OrganizationRole {
	m := &dbmodel.OrganizationRole{
		ID:             r.ID(),
		OrganizationID: r.OrganizationID().UUID(),
		Code:           r.Code(),
		Name:           r.Name(),
		Description:    r.Description(),
		BaseRole:       string(r.BaseRole()),
		CreatedAt:      r.CreatedAt(),
		UpdatedAt:      r.UpdatedAt(),
		DeletedAt:      r.DeletedAt(),
	}
	return m
}

func toDomainOrganizationRole(m *dbmodel.OrganizationRole) (*memberDomain.OrganizationRole, error) {
	orgID, err := orgDomain.OrganizationIDFromUUID(m.OrganizationID)
	if err != nil {
		return nil, err
	}
	return memberDomain.RehydrateOrganizationRole(memberDomain.RehydrateOrganizationRoleParams{
		ID:             m.ID,
		OrganizationID: orgID,
		Code:           m.Code,
		Name:           m.Name,
		Description:    m.Description,
		BaseRole:       m.BaseRole,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      m.DeletedAt,
	})
}
