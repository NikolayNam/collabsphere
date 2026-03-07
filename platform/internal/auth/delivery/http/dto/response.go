package dto

import "github.com/google/uuid"

type TokenResponse struct {
	Status int `json:"-"`
	Body   struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		TokenType    string `json:"tokenType"`
		ExpiresIn    int64  `json:"expiresIn"`
	}
}

type MeResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		FirstName string    `json:"firstName"`
		LastName  string    `json:"lastName"`
		Status    string    `json:"status"`
	}
}

type EmptyResponse struct {
	Status int `json:"-"`
}
