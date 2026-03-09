package http

import (
	"context"
	"strings"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/mapper"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
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

func (h *Handler) UpdateProductCategory(ctx context.Context, input *dto.UpdateProductCategoryInput) (*dto.ProductCategoryResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	categoryID, err := parseCategoryID(input.CategoryID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	parentID, err := parsePatchCategoryID(input.Body.ParentID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	category, err := h.svc.UpdateProductCategory(ctx, catalogapp.UpdateProductCategoryCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		CategoryID:     categoryID,
		ParentID:       parentID,
		Code:           input.Body.Code,
		Name:           input.Body.Name,
		SortOrder:      input.Body.SortOrder,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductCategoryResponse(category, 200), nil
}

func (h *Handler) DeleteProductCategory(ctx context.Context, input *dto.DeleteProductCategoryInput) (*dto.EmptyResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	categoryID, err := parseCategoryID(input.CategoryID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	if err := h.svc.DeleteProductCategory(ctx, catalogapp.DeleteProductCategoryCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		CategoryID:     categoryID,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
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

func (h *Handler) UpdateProduct(ctx context.Context, input *dto.UpdateProductInput) (*dto.ProductResponse, error) {
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
	categoryID, err := parsePatchCategoryID(input.Body.CategoryID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	product, err := h.svc.UpdateProduct(ctx, catalogapp.UpdateProductCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		ProductID:      productID,
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
	return mapper.ToProductResponse(product, 200), nil
}

func (h *Handler) DeleteProduct(ctx context.Context, input *dto.DeleteProductInput) (*dto.EmptyResponse, error) {
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

	if err := h.svc.DeleteProduct(ctx, catalogapp.DeleteProductCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		ProductID:      productID,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
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

func (h *Handler) CreateProductImportUpload(ctx context.Context, input *dto.CreateProductImportUploadInput) (*dto.ProductImportUploadResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	result, err := h.svc.CreateProductImportUpload(ctx, catalogapp.CreateProductImportUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductImportUploadResponse(result, 201), nil
}

func (h *Handler) RunProductImport(ctx context.Context, input *dto.RunProductImportInput) (*dto.ProductImportResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	sourceObjectID, err := parseUUID(input.Body.SourceObjectID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	view, err := h.svc.RunProductImport(ctx, catalogapp.RunProductImportCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		SourceObjectID: sourceObjectID,
		Mode:           input.Body.Mode,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductImportResponse(view, 200), nil
}

func (h *Handler) GetProductImport(ctx context.Context, input *dto.GetProductImportInput) (*dto.ProductImportResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	batchID, err := parseUUID(input.BatchID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}

	view, err := h.svc.GetProductImport(ctx, catalogapp.GetProductImportQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		BatchID:        batchID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductImportResponse(view, 200), nil
}

func currentActorAccountID(ctx context.Context) (accdomain.AccountID, error) {
	return httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
}

func parseOrganizationID(value string) (orgdomain.OrganizationID, error) {
	return httpbind.ParseOrganizationID(value, catalogapp.ErrValidation)
}

func parseOptionalCategoryID(value *string) (*catalogdomain.ProductCategoryID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	categoryID, err := parseCategoryID(trimmed)
	if err != nil {
		return nil, err
	}
	return &categoryID, nil
}

func parsePatchCategoryID(value *string) (*catalogdomain.ProductCategoryID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		zero := catalogdomain.ProductCategoryID{}
		return &zero, nil
	}
	categoryID, err := parseCategoryID(trimmed)
	if err != nil {
		return nil, err
	}
	return &categoryID, nil
}

func parseCategoryID(value string) (catalogdomain.ProductCategoryID, error) {
	id, err := parseUUID(value)
	if err != nil {
		return catalogdomain.ProductCategoryID{}, err
	}
	return catalogdomain.ProductCategoryIDFromUUID(id)
}

func parseProductID(value string) (catalogdomain.ProductID, error) {
	id, err := parseUUID(value)
	if err != nil {
		return catalogdomain.ProductID{}, err
	}
	return catalogdomain.ProductIDFromUUID(id)
}

func parseUUID(value string) (uuid.UUID, error) {
	return httpbind.ParseUUID(value, catalogapp.ErrValidation)
}
