package domain

import "github.com/google/uuid"

type AccountID uuid.UUID

func NewAccountID() AccountID {
	return AccountID(uuid.New())
}

func AccountIDFromUUID(id uuid.UUID) (AccountID, error) {
	if id == uuid.Nil {
		return AccountID{}, ErrUserIDEmpty
	}
	return AccountID(id), nil
}

func (id AccountID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id AccountID) String() string {
	return uuid.UUID(id).String()
}

func (id AccountID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}
