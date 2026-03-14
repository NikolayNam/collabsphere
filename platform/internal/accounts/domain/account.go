package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Account struct {
	id             AccountID
	email          Email
	passwordHash   PasswordHash
	zitadelUserID  *string
	displayName    *string
	avatarObjectID *uuid.UUID
	bio            *string
	phone          *string
	locale         *string
	timezone       *string
	website        *string
	isActive       bool
	createdAt      time.Time
	updatedAt      *time.Time
}

type NewAccountParams struct {
	ID           AccountID
	Email        Email
	PasswordHash PasswordHash
	DisplayName  *string
	Now          time.Time
}

func NewAccount(p NewAccountParams) (*Account, error) {
	if err := validateAccountCore(p.ID, p.Email); err != nil {
		return nil, err
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	displayName, err := normalizeOptionalDisplayName(p.DisplayName)
	if err != nil {
		return nil, err
	}

	updatedAt := p.Now

	return &Account{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		displayName:  displayName,
		isActive:     true,
		createdAt:    p.Now,
		updatedAt:    &updatedAt,
	}, nil
}

type RehydrateAccountParams struct {
	ID             AccountID
	Email          Email
	PasswordHash   PasswordHash
	ZitadelUserID  *string
	DisplayName    *string
	AvatarObjectID *uuid.UUID
	Bio            *string
	Phone          *string
	Locale         *string
	Timezone       *string
	Website        *string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func RehydrateAccount(p RehydrateAccountParams) (*Account, error) {
	if err := validateAccountCore(p.ID, p.Email); err != nil {
		return nil, err
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	displayName, err := normalizeOptionalDisplayName(p.DisplayName)
	if err != nil {
		return nil, err
	}
	bio, err := normalizeOptionalProfileField(p.Bio, 4096, ErrBioInvalid)
	if err != nil {
		return nil, err
	}
	phone, err := normalizeOptionalProfileField(p.Phone, 32, ErrPhoneInvalid)
	if err != nil {
		return nil, err
	}
	locale, err := normalizeOptionalProfileField(p.Locale, 16, ErrLocaleInvalid)
	if err != nil {
		return nil, err
	}
	timezone, err := normalizeOptionalProfileField(p.Timezone, 64, ErrTimezoneInvalid)
	if err != nil {
		return nil, err
	}
	website, err := normalizeOptionalProfileField(p.Website, 512, ErrWebsiteInvalid)
	if err != nil {
		return nil, err
	}

	updatedAt := p.UpdatedAt

	return &Account{
		id:             p.ID,
		email:          p.Email,
		passwordHash:   p.PasswordHash,
		zitadelUserID:  cloneStringPtr(p.ZitadelUserID),
		displayName:    displayName,
		avatarObjectID: cloneUUIDPtr(p.AvatarObjectID),
		bio:            bio,
		phone:          phone,
		locale:         locale,
		timezone:       timezone,
		website:        website,
		isActive:       p.IsActive,
		createdAt:      p.CreatedAt,
		updatedAt:      &updatedAt,
	}, nil
}

type AccountProfilePatch struct {
	DisplayName    *string
	AvatarObjectID *uuid.UUID
	ClearAvatar    bool
	Bio            *string
	Phone          *string
	Locale         *string
	Timezone       *string
	Website        *string
	UpdatedAt      time.Time
}

func (a *Account) ApplyProfilePatch(p AccountProfilePatch) error {
	if a == nil {
		return ErrUserIDEmpty
	}
	if p.UpdatedAt.IsZero() {
		return ErrNowRequired
	}
	if p.ClearAvatar {
		a.avatarObjectID = nil
	} else if p.AvatarObjectID != nil {
		a.avatarObjectID = cloneUUIDPtr(p.AvatarObjectID)
	}
	if p.DisplayName != nil {
		v, err := normalizeOptionalDisplayName(p.DisplayName)
		if err != nil {
			return err
		}
		a.displayName = v
	}
	if p.Bio != nil {
		v, err := normalizeOptionalProfileField(p.Bio, 4096, ErrBioInvalid)
		if err != nil {
			return err
		}
		a.bio = v
	}
	if p.Phone != nil {
		v, err := normalizeOptionalProfileField(p.Phone, 32, ErrPhoneInvalid)
		if err != nil {
			return err
		}
		a.phone = v
	}
	if p.Locale != nil {
		v, err := normalizeOptionalProfileField(p.Locale, 16, ErrLocaleInvalid)
		if err != nil {
			return err
		}
		a.locale = v
	}
	if p.Timezone != nil {
		v, err := normalizeOptionalProfileField(p.Timezone, 64, ErrTimezoneInvalid)
		if err != nil {
			return err
		}
		a.timezone = v
	}
	if p.Website != nil {
		v, err := normalizeOptionalProfileField(p.Website, 512, ErrWebsiteInvalid)
		if err != nil {
			return err
		}
		a.website = v
	}
	updatedAt := p.UpdatedAt
	a.updatedAt = &updatedAt
	return nil
}

func (a *Account) ID() AccountID              { return a.id }
func (a *Account) Email() Email               { return a.email }
func (a *Account) PasswordHash() PasswordHash { return a.passwordHash }
func (a *Account) ZitadelUserID() *string     { return cloneStringPtr(a.zitadelUserID) }
func (a *Account) DisplayName() *string       { return cloneStringPtr(a.displayName) }
func (a *Account) AvatarObjectID() *uuid.UUID { return cloneUUIDPtr(a.avatarObjectID) }
func (a *Account) Bio() *string               { return cloneStringPtr(a.bio) }
func (a *Account) Phone() *string             { return cloneStringPtr(a.phone) }
func (a *Account) Locale() *string            { return cloneStringPtr(a.locale) }
func (a *Account) Timezone() *string          { return cloneStringPtr(a.timezone) }
func (a *Account) Website() *string           { return cloneStringPtr(a.website) }
func (a *Account) IsActive() bool             { return a.isActive }

func (a *Account) Status() AccountStatus {
	if a.isActive {
		return AccountStatusActive
	}
	return AccountStatusBlocked
}

func (a *Account) CreatedAt() time.Time  { return a.createdAt }
func (a *Account) UpdatedAt() *time.Time { return cloneTimePtr(a.updatedAt) }

func validateAccountCore(id AccountID, email Email) error {
	switch {
	case id.IsZero():
		return ErrUserIDEmpty
	case email.IsZero():
		return ErrEmailEmpty
	default:
		return nil
	}
}

func normalizeOptionalDisplayName(s *string) (*string, error) {
	return normalizeOptionalProfileField(s, 255, ErrDisplayNameInvalid)
}

func normalizeOptionalProfileField(s *string, maxRunes int, invalid error) (*string, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(v) > maxRunes {
		return nil, invalid
	}
	return &v, nil
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}

func cloneUUIDPtr(id *uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	v := *id
	return &v
}
