package dto

import "github.com/google/uuid"

type AccountResponse struct {
	Body struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Status    string    `json:"status"`
	}
}
