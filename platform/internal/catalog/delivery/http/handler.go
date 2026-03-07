package http

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/mapper"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	svc *catalogapp.Service
}

func NewHandler(svc *catalogapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateProductCategory(ctx context.Context, input *dto.CreateProductCategoryInput) (*dto.ProductCategoryResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	parentID, err := parseOptionalCategoryID(input.Body.ParentID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	category, err := h.svc.CreateProductCategory(ctx, catalogapp.CreateProductCategoryCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		ParentID:       parentID,
		Code:           input.Body.Code,
		Name:           input.Body.Name,
		SortOrder:      input.Body.SortOrder,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductCategoryResponse(category, 201), nil
}

func (h *Handler) ListProductCategories(ctx context.Context, input *dto.ListProductCategoriesInput) (*dto.ProductCategoriesResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	categories, err := h.svc.ListProductCategories(ctx, catalogapp.ListProductCategoriesQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductCategoriesResponse(categories), nil
}

func (h *Handler) CreateProduct(ctx context.Context, input *dto.CreateProductInput) (*dto.ProductResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	categoryID, err := parseOptionalCategoryID(input.Body.CategoryID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	product, err := h.svc.CreateProduct(ctx, catalogapp.CreateProductCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		CategoryID:     categoryID,
		Name:           input.Body.Name,
		Description:    input.Body.Description,
		SKU:            input.Body.SKU,
		PriceAmount:    input.Body.PriceAmount,
		CurrencyCode:   input.Body.CurrencyCode,
		IsActive:       input.Body.IsActive,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductResponse(product, 201), nil
}

func (h *Handler) ListProducts(ctx context.Context, input *dto.ListProductsInput) (*dto.ProductsResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	products, err := h.svc.ListProducts(ctx, catalogapp.ListProductsQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductsResponse(products), nil
}

func (h *Handler) GetProductByID(ctx context.Context, input *dto.GetProductByIDInput) (*dto.ProductResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	productID, err := parseProductID(input.ProductID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	product, err := h.svc.GetProductByID(ctx, catalogapp.GetProductByIDQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		ProductID:      productID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductResponse(product, 200), nil
}

func currentActorAccountID(ctx context.Context) (accdomain.AccountID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.Authenticated {
		return accdomain.AccountID{}, fault.Unauthorized("Authentication required")
	}
	accountID, err := accdomain.AccountIDFromUUID(principal.AccountID)
	if err != nil {
		return accdomain.AccountID{}, fault.Unauthorized("Authentication required")
	}
	return accountID, nil
}

func parseOrganizationID(value string) (orgdomain.OrganizationID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return orgdomain.OrganizationID{}, catalogapp.ErrValidation
	}
	return orgdomain.OrganizationIDFromUUID(id)
}

func parseOptionalCategoryID(value *string) (*catalogdomain.ProductCategoryID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := *value
	if trimmed == "" {
		return nil, nil
	}
	categoryID, err := parseCategoryID(trimmed)
	if err != nil {
		return nil, err
	}
	return &categoryID, nil
}

func parseCategoryID(value string) (catalogdomain.ProductCategoryID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return catalogdomain.ProductCategoryID{}, catalogapp.ErrValidation
	}
	return catalogdomain.ProductCategoryIDFromUUID(id)
}

func parseProductID(value string) (catalogdomain.ProductID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return catalogdomain.ProductID{}, catalogapp.ErrValidation
	}
	return catalogdomain.ProductIDFromUUID(id)
}
