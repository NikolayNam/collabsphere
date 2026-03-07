package domain

import "github.com/google/uuid"

type ProductCategoryID uuid.UUID

func NewProductCategoryID() ProductCategoryID {
	return ProductCategoryID(uuid.New())
}

func ProductCategoryIDFromUUID(id uuid.UUID) (ProductCategoryID, error) {
	if id == uuid.Nil {
		return ProductCategoryID{}, ErrProductCategoryIDEmpty
	}
	return ProductCategoryID(id), nil
}

func (id ProductCategoryID) UUID() uuid.UUID { return uuid.UUID(id) }
func (id ProductCategoryID) String() string  { return uuid.UUID(id).String() }
func (id ProductCategoryID) IsZero() bool    { return uuid.UUID(id) == uuid.Nil }
