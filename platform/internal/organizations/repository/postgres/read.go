package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
)

func (r *OrganizationRepo) GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	var m dbmodel.Organization

	err := r.dbFrom(ctx).WithContext(ctx).
		Take(&m, "id = ?", id.UUID()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainOrganization(&m)
}

func (r *OrganizationRepo) Exists(ctx context.Context, id domain.OrganizationID) (bool, error) {
	if id.IsZero() {
		return false, nil
	}

	var n int64
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organizations").
		Where("id = ?", id.UUID()).
		Limit(1).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
