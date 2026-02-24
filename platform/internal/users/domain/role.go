package domain

import "errors"

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func NewRole(raw string) (Role, error) {
	r := Role(raw)
	if !r.Valid() {
		return "", errors.New("role is invalid")
	}
	return r, nil
}
