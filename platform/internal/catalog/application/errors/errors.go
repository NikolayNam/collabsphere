package errors

import "github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"

var (
	ErrValidation = fault.ErrValidation
	ErrConflict   = fault.ErrConflict
	ErrInternal   = fault.ErrInternal
	ErrNotFound   = fault.ErrNotFound
	ErrForbidden  = fault.ErrForbidden
)

const (
	CodeInvalidInput            = "CATALOG_INVALID_INPUT"
	CodeOrganizationNotFound    = "CATALOG_ORGANIZATION_NOT_FOUND"
	CodeProductCategoryExists   = "PRODUCT_CATEGORY_ALREADY_EXISTS"
	CodeProductCategoryNotFound = "PRODUCT_CATEGORY_NOT_FOUND"
	CodeProductExists           = "PRODUCT_ALREADY_EXISTS"
	CodeProductNotFound         = "PRODUCT_NOT_FOUND"
	CodeSourceObjectNotFound    = "PRODUCT_IMPORT_SOURCE_OBJECT_NOT_FOUND"
	CodeProductImportNotFound   = "PRODUCT_IMPORT_NOT_FOUND"
	CodeImportUnavailable       = "PRODUCT_IMPORT_UNAVAILABLE"
	CodeImportFileInvalid       = "PRODUCT_IMPORT_FILE_INVALID"
	CodeAccessDenied            = "CATALOG_ACCESS_DENIED"
	CodeInternal                = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

func OrganizationNotFound() error {
	return fault.NotFound("Organization not found", fault.Code(CodeOrganizationNotFound))
}

func ProductCategoryAlreadyExists() error {
	return fault.Conflict("Product category already exists", fault.Code(CodeProductCategoryExists))
}

func ProductCategoryNotFound() error {
	return fault.NotFound("Product category not found", fault.Code(CodeProductCategoryNotFound))
}

func ProductAlreadyExists() error {
	return fault.Conflict("Product already exists", fault.Code(CodeProductExists))
}

func ProductNotFound() error {
	return fault.NotFound("Product not found", fault.Code(CodeProductNotFound))
}

func ProductImportSourceObjectNotFound() error {
	return fault.NotFound("Import source object not found", fault.Code(CodeSourceObjectNotFound))
}

func ProductImportNotFound() error {
	return fault.NotFound("Product import batch not found", fault.Code(CodeProductImportNotFound))
}

func ProductImportUnavailable() error {
	return fault.Unavailable("Product import storage is unavailable", fault.Code(CodeImportUnavailable))
}

func ProductImportFileInvalid(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeImportFileInvalid)}, opts...)
	return fault.Validation(message, opts...)
}

func AccessDenied() error {
	return fault.Forbidden("Access denied", fault.Code(CodeAccessDenied))
}

func Internal(detail string, cause error) error {
	_ = detail
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}
