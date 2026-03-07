package postgres

import (
	"context"
	"fmt"
	"time"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type productCategoryTemplateRow struct {
	ID        uuid.UUID  `gorm:"column:id"`
	ParentID  *uuid.UUID `gorm:"column:parent_id"`
	Code      string     `gorm:"column:code"`
	Name      string     `gorm:"column:name"`
	SortOrder int64      `gorm:"column:sort_order"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

type existingProductCategoryRow struct {
	ID         uuid.UUID  `gorm:"column:id"`
	TemplateID *uuid.UUID `gorm:"column:template_id"`
}

type productCategoryRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	ParentID       *uuid.UUID `gorm:"column:parent_id"`
	TemplateID     *uuid.UUID `gorm:"column:template_id"`
	Code           string     `gorm:"column:code"`
	Name           string     `gorm:"column:name"`
	SortOrder      int64      `gorm:"column:sort_order"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (r *ProductCategoryRepo) ProvisionDefaults(ctx context.Context, organizationID orgdomain.OrganizationID, now time.Time) error {
	if organizationID.IsZero() {
		return fmt.Errorf("organization id is required")
	}
	if now.IsZero() {
		return fmt.Errorf("provision timestamp is required")
	}

	db := r.dbFrom(ctx).WithContext(ctx)

	var templates []productCategoryTemplateRow
	if err := db.Table("catalog.product_category_templates").
		Select("id", "parent_id", "code", "name", "sort_order", "deleted_at").
		Order("sort_order ASC, id ASC").
		Find(&templates).Error; err != nil {
		return err
	}
	if len(templates) == 0 {
		return nil
	}

	var existing []existingProductCategoryRow
	if err := db.Table("catalog.product_categories").
		Select("id", "template_id").
		Where("organization_id = ? AND template_id IS NOT NULL", organizationID.UUID()).
		Find(&existing).Error; err != nil {
		return err
	}

	createdByTemplate := make(map[uuid.UUID]uuid.UUID, len(existing)+len(templates))
	for _, row := range existing {
		if row.TemplateID != nil {
			createdByTemplate[*row.TemplateID] = row.ID
		}
	}

	pending := make(map[uuid.UUID]productCategoryTemplateRow, len(templates))
	for _, tmpl := range templates {
		if _, ok := createdByTemplate[tmpl.ID]; ok {
			continue
		}
		pending[tmpl.ID] = tmpl
	}

	for len(pending) > 0 {
		progressed := false

		for templateID, tmpl := range pending {
			var parentID *uuid.UUID
			if tmpl.ParentID != nil {
				resolvedParentID, ok := createdByTemplate[*tmpl.ParentID]
				if !ok {
					continue
				}
				parentCopy := resolvedParentID
				parentID = &parentCopy
			}

			templateCopy := templateID
			row := productCategoryRow{
				ID:             uuid.New(),
				OrganizationID: organizationID.UUID(),
				ParentID:       parentID,
				TemplateID:     &templateCopy,
				Code:           tmpl.Code,
				Name:           tmpl.Name,
				SortOrder:      tmpl.SortOrder,
				CreatedAt:      now,
				UpdatedAt:      now,
				DeletedAt:      tmpl.DeletedAt,
			}

			if err := db.Table("catalog.product_categories").Create(&row).Error; err != nil {
				return err
			}

			createdByTemplate[templateID] = row.ID
			delete(pending, templateID)
			progressed = true
		}

		if !progressed {
			return fmt.Errorf("catalog product category templates contain unresolved parent links")
		}
	}

	return nil
}
