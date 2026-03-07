package domain

import "github.com/google/uuid"

type Principal struct {
	AccountID     uuid.UUID
	SessionID     uuid.UUID
	Authenticated bool
}

func AnonymousPrincipal() Principal {
	return Principal{}
}

func NewPrincipal(accountID, sessionID uuid.UUID) Principal {
	return Principal{
		AccountID:     accountID,
		SessionID:     sessionID,
		Authenticated: accountID != uuid.Nil,
	}
}
