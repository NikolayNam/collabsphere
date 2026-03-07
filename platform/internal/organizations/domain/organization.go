package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Organization struct {
	id OrganizationID

	legalName   string
	displayName *string

	primaryEmail Email

	status OrganizationStatus

	createdAt time.Time
	updatedAt *time.Time
}

type NewOrganizationParams struct {
	ID           OrganizationID
	LegalName    string
	DisplayName  *string
	PrimaryEmail Email
	Now          time.Time
}

func NewOrganization(p NewOrganizationParams) (*Organization, error) {
	if p.ID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.PrimaryEmail.IsZero() {
		return nil, ErrEmailEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	ln, err := normalizeLegalName(p.LegalName)
	if err != nil {
		return nil, err
	}

	dn, err := normalizeOptionalDisplayName(p.DisplayName)
	if err != nil {
		return nil, err
	}

	return &Organization{
		id:           p.ID,
		legalName:    ln,
		displayName:  dn,
		primaryEmail: p.PrimaryEmail,
		status:       OrganizationStatusActive,
		createdAt:    p.Now,
		updatedAt:    nil,
	}, nil
}

type RehydrateOrganizationParams struct {
	ID           OrganizationID
	LegalName    string
	DisplayName  *string
	PrimaryEmail Email
	Status       OrganizationStatus
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func RehydrateOrganization(p RehydrateOrganizationParams) (*Organization, error) {
	if p.ID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.PrimaryEmail.IsZero() {
		return nil, ErrEmailEmpty
	}
	if !p.Status.IsValid() {
		return nil, ErrInvalidOrganizationStatus
	}
	if p.CreatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt != nil && p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	ln, err := normalizeLegalName(p.LegalName)
	if err != nil {
		return nil, err
	}
	dn, err := normalizeOptionalDisplayName(p.DisplayName)
	if err != nil {
		return nil, err
	}

	return &Organization{
		id:           p.ID,
		legalName:    ln,
		displayName:  dn,
		primaryEmail: p.PrimaryEmail,
		status:       p.Status,
		createdAt:    p.CreatedAt,
		updatedAt:    cloneTimePtr(p.UpdatedAt),
	}, nil
}

// Accessors

func (t *Organization) ID() OrganizationID         { return t.id }
func (t *Organization) LegalName() string          { return t.legalName }
func (t *Organization) DisplayName() *string       { return cloneStringPtr(t.displayName) }
func (t *Organization) PrimaryEmail() Email        { return t.primaryEmail }
func (t *Organization) Status() OrganizationStatus { return t.status }
func (t *Organization) CreatedAt() time.Time       { return t.createdAt }
func (t *Organization) UpdatedAt() *time.Time      { return cloneTimePtr(t.updatedAt) }

// Helpers

func normalizeLegalName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 500 {
		return "", ErrLegalNameInvalid
	}
	return s, nil
}

func normalizeOptionalDisplayName(s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(v) > 200 {
		return nil, ErrDisplayNameInvalid
	}
	return &v, nil
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return new(*t)
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	return new(*s)
}
