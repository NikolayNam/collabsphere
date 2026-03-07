package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProductCategoryBody struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	ParentID       *uuid.UUID `json:"parentId,omitempty"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	SortOrder      int64      `json:"sortOrder"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type ProductBody struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	CategoryID     *uuid.UUID `json:"categoryId,omitempty"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	SKU            *string    `json:"sku,omitempty"`
	PriceAmount    *string    `json:"priceAmount,omitempty"`
	CurrencyCode   *string    `json:"currencyCode,omitempty"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type ProductCategoryResponse struct {
	Status int                 `json:"-"`
	Body   ProductCategoryBody `json:"body"`
}

type ProductCategoriesResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []ProductCategoryBody `json:"items"`
	} `json:"body"`
}

type ProductResponse struct {
	Status int         `json:"-"`
	Body   ProductBody `json:"body"`
}

type ProductsResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []ProductBody `json:"items"`
	} `json:"body"`
}
