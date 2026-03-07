package domain

import "github.com/google/uuid"

type ProductID uuid.UUID

func NewProductID() ProductID {
	return ProductID(uuid.New())
}

func ProductIDFromUUID(id uuid.UUID) (ProductID, error) {
	if id == uuid.Nil {
		return ProductID{}, ErrProductIDEmpty
	}
	return ProductID(id), nil
}

func (id ProductID) UUID() uuid.UUID { return uuid.UUID(id) }
func (id ProductID) String() string  { return uuid.UUID(id).String() }
func (id ProductID) IsZero() bool    { return uuid.UUID(id) == uuid.Nil }
