package domain

import (
	"errors"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func UserIDFromUUID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, errors.New("user id is nil")
	}
	return UserID(id), nil
}

func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id UserID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}
