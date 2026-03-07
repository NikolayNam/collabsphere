package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
)

func (r *OrganizationRepo) Create(ctx context.Context, organization *domain.Organization) error {
	if organization == nil {
		return apperrors.InvalidInput("Organization is required")
	}

	m := mapper.ToDBOrganizationForCreate(organization)
	if m == nil {
		return errors.New("db organization model is nil")
	}

	err := r.dbFrom(ctx).WithContext(ctx).Create(m).Error
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		return err
	}
	return nil
}
