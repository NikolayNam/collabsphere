package domain

import "github.com/google/uuid"

type OrganizationID uuid.UUID

func NewOrganizationID() OrganizationID { return OrganizationID(uuid.New()) }

func OrganizationIDFromUUID(id uuid.UUID) (OrganizationID, error) {
	if id == uuid.Nil {
		return OrganizationID{}, ErrOrganizationIDEmpty
	}
	return OrganizationID(id), nil
}

func (id OrganizationID) UUID() uuid.UUID { return uuid.UUID(id) }
func (id OrganizationID) String() string  { return uuid.UUID(id).String() }
func (id OrganizationID) IsZero() bool    { return uuid.UUID(id) == uuid.Nil }
