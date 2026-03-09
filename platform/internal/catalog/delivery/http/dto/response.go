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

type ProductImportErrorBody struct {
	ID        uuid.UUID      `json:"id"`
	RowNo     *int           `json:"rowNo,omitempty"`
	Code      *string        `json:"code,omitempty"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type ProductImportBatchBody struct {
	ID                 uuid.UUID                `json:"id"`
	OrganizationID     uuid.UUID                `json:"organizationId"`
	SourceObjectID     uuid.UUID                `json:"sourceObjectId"`
	CreatedByAccountID uuid.UUID                `json:"createdByAccountId"`
	Status             string                   `json:"status"`
	TotalRows          *int                     `json:"totalRows,omitempty"`
	ProcessedRows      int                      `json:"processedRows"`
	SuccessRows        int                      `json:"successRows"`
	ErrorRows          int                      `json:"errorRows"`
	StartedBy          *string                  `json:"startedBy,omitempty"`
	StartedAt          time.Time                `json:"startedAt"`
	FinishedAt         *time.Time               `json:"finishedAt,omitempty"`
	CreatedAt          time.Time                `json:"createdAt"`
	UpdatedAt          *time.Time               `json:"updatedAt,omitempty"`
	Mode               *string                  `json:"mode,omitempty"`
	ResultSummary      map[string]any           `json:"resultSummary,omitempty"`
	Errors             []ProductImportErrorBody `json:"errors,omitempty"`
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

type ProductImportResponse struct {
	Status int                    `json:"-"`
	Body   ProductImportBatchBody `json:"body"`
}
