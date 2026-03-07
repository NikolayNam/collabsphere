package domain

import "github.com/google/uuid"

type GroupID uuid.UUID

func NewGroupID() GroupID {
	return GroupID(uuid.New())
}

func GroupIDFromUUID(id uuid.UUID) (GroupID, error) {
	if id == uuid.Nil {
		return GroupID{}, ErrGroupIDEmpty
	}
	return GroupID(id), nil
}

func (id GroupID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id GroupID) String() string {
	return uuid.UUID(id).String()
}

func (id GroupID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}
