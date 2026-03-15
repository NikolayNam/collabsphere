package domain

import "errors"

var (
	ErrProductCategoryIDEmpty       = errors.New("product category id is required")
	ErrProductCategoryCodeInvalid   = errors.New("product category code is invalid")
	ErrProductCategoryNameInvalid   = errors.New("product category name is invalid")
	ErrProductCategorySortInvalid   = errors.New("product category sort order is invalid")
	ErrProductCategoryStatusInvalid = errors.New("product category status is invalid")
	ErrProductIDEmpty               = errors.New("product id is required")
	ErrProductNameInvalid           = errors.New("product name is invalid")
	ErrProductPriceInvalid          = errors.New("product price is invalid")
	ErrProductCurrencyInvalid       = errors.New("product currency is invalid")
	ErrProductSKUInvalid            = errors.New("product sku is invalid")
	ErrProductStatusInvalid         = errors.New("product status is invalid")
	ErrNowRequired                  = errors.New("current time is required")
	ErrTimestampsMissing            = errors.New("timestamps are required")
	ErrTimestampsInvalid            = errors.New("timestamps are invalid")
)
