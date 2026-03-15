package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Number         string
	Title          string
	Description    *string
	Status         string
	BudgetAmount   *string
	CurrencyCode   *string
	CreatedAt      time.Time
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	CategoryID  *uuid.UUID
	ProductName *string
	Quantity    *string
	Unit        *string
	Note        *string
	SortOrder   int
}

type Offer struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	OrganizationID uuid.UUID
	Status         string
	Comment        *string
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
}

type OfferItem struct {
	ID           uuid.UUID
	OfferID      uuid.UUID
	CategoryID   *uuid.UUID
	ProductID    *uuid.UUID
	CustomTitle  *string
	Quantity     *string
	Unit         *string
	PriceAmount  *string
	CurrencyCode *string
	Note         *string
	SortOrder    int
}

type BoardComment struct {
	ID               uuid.UUID
	OrderID          *uuid.UUID
	OfferID          *uuid.UUID
	OrganizationID   uuid.UUID
	AccountID        uuid.UUID
	Comment          string
	OrganizationName *string
	CreatedAt        time.Time
}
