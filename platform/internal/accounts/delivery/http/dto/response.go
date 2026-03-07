package dto

import "github.com/google/uuid"

type AccountResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"firstName"`
		LastName  string    `json:"lastName"`
		Status    string    `json:"status"`
	}
}
