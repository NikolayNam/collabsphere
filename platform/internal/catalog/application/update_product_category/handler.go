package update_product_category

import (
	"context"
	stdErrors "errors"
	"strings"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	roleResolver  memberports.RoleResolver
	clock         ports.Clock
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, roleResolver memberports.RoleResolver, clock ports.Clock) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships, roleResolver: roleResolver, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*catalogdomain.ProductCategory, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, h.roleResolver, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}

	current, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, cmd.CategoryID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, catalogerrors.ProductCategoryNotFound()
	}

	parentID := current.ParentID()
	if cmd.ParentID != nil {
		if !cmd.ParentID.IsZero() {
			parent, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, *cmd.ParentID)
			if err != nil {
				return nil, err
			}
			if parent == nil {
				return nil, catalogerrors.ProductCategoryNotFound()
			}
		}
		parentID = cmd.ParentID
	}
	if parentID != nil && parentID.UUID() == current.ID().UUID() {
		return nil, catalogerrors.InvalidInput("Category parent cannot reference itself")
	}

	code := current.Code()
	if cmd.Code != nil {
		code = strings.TrimSpace(*cmd.Code)
	}
	name := current.Name()
	if cmd.Name != nil {
		name = strings.TrimSpace(*cmd.Name)
	}
	status := string(current.Status())
	if cmd.Status != nil {
		status = strings.TrimSpace(*cmd.Status)
	}
	sortOrder := current.SortOrder()
	if cmd.SortOrder != nil {
		sortOrder = *cmd.SortOrder
	}

	updated, err := catalogdomain.RehydrateProductCategory(catalogdomain.RehydrateProductCategoryParams{
		ID:             current.ID(),
		OrganizationID: current.OrganizationID(),
		ParentID:       parentID,
		TemplateID:     current.TemplateID(),
		Status:         status,
		Code:           code,
		Name:           name,
		SortOrder:      sortOrder,
		CreatedAt:      current.CreatedAt(),
		UpdatedAt:      h.clock.Now(),
	})
	if err != nil {
		return nil, catalogerrors.InvalidInput("Invalid product category data")
	}

	if err := h.repo.UpdateProductCategory(ctx, updated); err != nil {
		if stdErrors.Is(err, catalogerrors.ErrConflict) {
			return nil, catalogerrors.ProductCategoryAlreadyExists()
		}
		return nil, err
	}
	return updated, nil
}
