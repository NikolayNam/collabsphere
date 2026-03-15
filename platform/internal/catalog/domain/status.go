package domain

import "strings"

type ProductStatus string
type ProductCategoryStatus string

const (
	ProductStatusDraft      ProductStatus = "draft"
	ProductStatusValidating ProductStatus = "validating"
	ProductStatusVerified   ProductStatus = "verified"
	ProductStatusPublished  ProductStatus = "published"
	ProductStatusWithdrawn  ProductStatus = "withdrawn"
	ProductStatusArchived   ProductStatus = "archived"
)

const (
	ProductCategoryStatusDraft      ProductCategoryStatus = "draft"
	ProductCategoryStatusValidating ProductCategoryStatus = "validating"
	ProductCategoryStatusVerified   ProductCategoryStatus = "verified"
	ProductCategoryStatusPublished  ProductCategoryStatus = "published"
	ProductCategoryStatusWithdrawn  ProductCategoryStatus = "withdrawn"
	ProductCategoryStatusArchived   ProductCategoryStatus = "archived"
)

func normalizeProductStatus(value string) (ProductStatus, error) {
	switch ProductStatus(strings.ToLower(strings.TrimSpace(value))) {
	case ProductStatusDraft,
		ProductStatusValidating,
		ProductStatusVerified,
		ProductStatusPublished,
		ProductStatusWithdrawn,
		ProductStatusArchived:
		return ProductStatus(strings.ToLower(strings.TrimSpace(value))), nil
	default:
		return "", ErrProductStatusInvalid
	}
}

func normalizeProductStatusOrDefault(value string) (ProductStatus, error) {
	if strings.TrimSpace(value) == "" {
		return ProductStatusDraft, nil
	}
	return normalizeProductStatus(value)
}

func normalizeCategoryStatus(value string) (ProductCategoryStatus, error) {
	switch ProductCategoryStatus(strings.ToLower(strings.TrimSpace(value))) {
	case ProductCategoryStatusDraft,
		ProductCategoryStatusValidating,
		ProductCategoryStatusVerified,
		ProductCategoryStatusPublished,
		ProductCategoryStatusWithdrawn,
		ProductCategoryStatusArchived:
		return ProductCategoryStatus(strings.ToLower(strings.TrimSpace(value))), nil
	default:
		return "", ErrProductCategoryStatusInvalid
	}
}

func normalizeCategoryStatusOrDefault(value string) (ProductCategoryStatus, error) {
	if strings.TrimSpace(value) == "" {
		return ProductCategoryStatusDraft, nil
	}
	return normalizeCategoryStatus(value)
}
