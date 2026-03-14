package domain

import "github.com/google/uuid"

type SubjectType string

const (
	SubjectTypeUnknown SubjectType = ""
	SubjectTypeAccount SubjectType = "account"
	SubjectTypeGuest   SubjectType = "guest"
	SubjectTypeService SubjectType = "service"
)

type Principal struct {
	SubjectType   SubjectType
	SubjectID     uuid.UUID
	AccountID     uuid.UUID
	GuestID       uuid.UUID
	ServiceID     uuid.UUID
	SessionID     uuid.UUID
	ChannelID     uuid.UUID
	Authenticated bool
}

func AnonymousPrincipal() Principal {
	return Principal{}
}

func NewPrincipal(accountID, sessionID uuid.UUID) Principal {
	return NewAccountPrincipal(accountID, sessionID)
}

func NewAccountPrincipal(accountID, sessionID uuid.UUID) Principal {
	return Principal{
		SubjectType:   SubjectTypeAccount,
		SubjectID:     accountID,
		AccountID:     accountID,
		SessionID:     sessionID,
		Authenticated: accountID != uuid.Nil,
	}
}

func NewGuestPrincipal(guestID, sessionID, channelID uuid.UUID) Principal {
	return Principal{
		SubjectType:   SubjectTypeGuest,
		SubjectID:     guestID,
		GuestID:       guestID,
		SessionID:     sessionID,
		ChannelID:     channelID,
		Authenticated: guestID != uuid.Nil,
	}
}

func NewServicePrincipal(serviceID, sessionID uuid.UUID) Principal {
	return Principal{
		SubjectType:   SubjectTypeService,
		SubjectID:     serviceID,
		ServiceID:     serviceID,
		SessionID:     sessionID,
		Authenticated: serviceID != uuid.Nil,
	}
}

func (p Principal) IsAccount() bool {
	return p.Authenticated && p.SubjectType == SubjectTypeAccount && p.AccountID != uuid.Nil
}

func (p Principal) IsGuest() bool {
	return p.Authenticated && p.SubjectType == SubjectTypeGuest && p.GuestID != uuid.Nil
}

func (p Principal) IsService() bool {
	return p.Authenticated && p.SubjectType == SubjectTypeService && p.ServiceID != uuid.Nil
}
