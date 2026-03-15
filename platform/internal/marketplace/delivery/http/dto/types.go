package dto

import "time"

type OrderItemPayload struct {
	ID          string  `json:"id"`
	CategoryID  *string `json:"categoryId,omitempty"`
	ProductName *string `json:"productName,omitempty"`
	Quantity    *string `json:"quantity,omitempty"`
	Unit        *string `json:"unit,omitempty"`
	Note        *string `json:"note,omitempty"`
	SortOrder   int     `json:"sortOrder"`
}

type BoardCommentPayload struct {
	ID               string  `json:"id"`
	OrderID          *string `json:"orderId,omitempty"`
	OfferID          *string `json:"offerId,omitempty"`
	OrganizationID   string  `json:"organizationId"`
	OrganizationName *string `json:"organizationName,omitempty"`
	AccountID        string  `json:"accountId"`
	Comment          string  `json:"comment"`
	CreatedAt        string  `json:"createdAt"`
}

type OrderPayload struct {
	ID             string                `json:"id"`
	OrganizationID string                `json:"organizationId"`
	Number         string                `json:"number"`
	Title          string                `json:"title"`
	Description    *string               `json:"description,omitempty"`
	Status         string                `json:"status"`
	BudgetAmount   *string               `json:"budgetAmount,omitempty"`
	CurrencyCode   *string               `json:"currencyCode,omitempty"`
	CreatedAt      string                `json:"createdAt"`
	Items          []OrderItemPayload    `json:"items,omitempty"`
	Comments       []BoardCommentPayload `json:"comments,omitempty"`
}

type OfferItemPayload struct {
	ID           string  `json:"id"`
	CategoryID   *string `json:"categoryId,omitempty"`
	ProductID    *string `json:"productId,omitempty"`
	CustomTitle  *string `json:"customTitle,omitempty"`
	Quantity     *string `json:"quantity,omitempty"`
	Unit         *string `json:"unit,omitempty"`
	PriceAmount  *string `json:"priceAmount,omitempty"`
	CurrencyCode *string `json:"currencyCode,omitempty"`
	Note         *string `json:"note,omitempty"`
	SortOrder    int     `json:"sortOrder"`
}

type OfferPayload struct {
	ID             string                `json:"id"`
	OrderID        string                `json:"orderId"`
	OrganizationID string                `json:"organizationId"`
	Status         string                `json:"status"`
	Comment        *string               `json:"comment,omitempty"`
	CreatedBy      *string               `json:"createdBy,omitempty"`
	CreatedAt      string                `json:"createdAt"`
	Items          []OfferItemPayload    `json:"items,omitempty"`
	Comments       []BoardCommentPayload `json:"comments,omitempty"`
}

type OrdersListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []OrderPayload `json:"items"`
	} `json:"body"`
}

type OrderResponse struct {
	Status int          `json:"-"`
	Body   OrderPayload `json:"body"`
}

type OffersListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []OfferPayload `json:"items"`
	} `json:"body"`
}

type OfferResponse struct {
	Status int          `json:"-"`
	Body   OfferPayload `json:"body"`
}

type EmptyResponse struct {
	Status int `json:"-"`
}

func FormatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
