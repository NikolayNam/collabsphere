package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Organization struct {
	id           OrganizationID
	name         string
	slug         string
	logoObjectID *uuid.UUID
	description  *string
	website      *string
	primaryEmail *Email
	phone        *string
	address      *string
	industry     *string
	isActive     bool
	createdAt    time.Time
	updatedAt    *time.Time
}

type NewOrganizationParams struct {
	ID   OrganizationID
	Name string
	Slug string
	Now  time.Time
}

func NewOrganization(p NewOrganizationParams) (*Organization, error) {
	if p.ID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	name, err := normalizeName(p.Name)
	if err != nil {
		return nil, err
	}
	slug, err := normalizeSlug(p.Slug)
	if err != nil {
		return nil, err
	}

	updatedAt := p.Now

	return &Organization{
		id:        p.ID,
		name:      name,
		slug:      slug,
		isActive:  true,
		createdAt: p.Now,
		updatedAt: &updatedAt,
	}, nil
}

type RehydrateOrganizationParams struct {
	ID           OrganizationID
	Name         string
	Slug         string
	LogoObjectID *uuid.UUID
	Description  *string
	Website      *string
	PrimaryEmail *string
	Phone        *string
	Address      *string
	Industry     *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func RehydrateOrganization(p RehydrateOrganizationParams) (*Organization, error) {
	if p.ID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	name, err := normalizeName(p.Name)
	if err != nil {
		return nil, err
	}
	slug, err := normalizeSlug(p.Slug)
	if err != nil {
		return nil, err
	}
	description, err := normalizeOptionalOrgField(p.Description, 4096, ErrDescriptionInvalid)
	if err != nil {
		return nil, err
	}
	website, err := normalizeOptionalOrgField(p.Website, 512, ErrWebsiteInvalid)
	if err != nil {
		return nil, err
	}
	primaryEmail, err := normalizeOptionalOrganizationEmail(p.PrimaryEmail)
	if err != nil {
		return nil, err
	}
	phone, err := normalizeOptionalOrgField(p.Phone, 32, ErrPhoneInvalid)
	if err != nil {
		return nil, err
	}
	address, err := normalizeOptionalOrgField(p.Address, 4096, ErrAddressInvalid)
	if err != nil {
		return nil, err
	}
	industry, err := normalizeOptionalOrgField(p.Industry, 128, ErrIndustryInvalid)
	if err != nil {
		return nil, err
	}

	updatedAt := p.UpdatedAt

	return &Organization{
		id:           p.ID,
		name:         name,
		slug:         slug,
		logoObjectID: cloneUUIDPtr(p.LogoObjectID),
		description:  description,
		website:      website,
		primaryEmail: primaryEmail,
		phone:        phone,
		address:      address,
		industry:     industry,
		isActive:     p.IsActive,
		createdAt:    p.CreatedAt,
		updatedAt:    &updatedAt,
	}, nil
}

type OrganizationProfilePatch struct {
	Name         *string
	Slug         *string
	LogoObjectID *uuid.UUID
	ClearLogo    bool
	Description  *string
	Website      *string
	PrimaryEmail *string
	Phone        *string
	Address      *string
	Industry     *string
	UpdatedAt    time.Time
}

func (o *Organization) ApplyProfilePatch(p OrganizationProfilePatch) error {
	if o == nil {
		return ErrOrganizationIDEmpty
	}
	if p.UpdatedAt.IsZero() {
		return ErrNowRequired
	}
	if p.Name != nil {
		name, err := normalizeName(*p.Name)
		if err != nil {
			return err
		}
		o.name = name
	}
	if p.Slug != nil {
		slug, err := normalizeSlug(*p.Slug)
		if err != nil {
			return err
		}
		o.slug = slug
	}
	if p.ClearLogo {
		o.logoObjectID = nil
	} else if p.LogoObjectID != nil {
		o.logoObjectID = cloneUUIDPtr(p.LogoObjectID)
	}
	if p.Description != nil {
		v, err := normalizeOptionalOrgField(p.Description, 4096, ErrDescriptionInvalid)
		if err != nil {
			return err
		}
		o.description = v
	}
	if p.Website != nil {
		v, err := normalizeOptionalOrgField(p.Website, 512, ErrWebsiteInvalid)
		if err != nil {
			return err
		}
		o.website = v
	}
	if p.PrimaryEmail != nil {
		v, err := normalizeOptionalOrganizationEmail(p.PrimaryEmail)
		if err != nil {
			return err
		}
		o.primaryEmail = v
	}
	if p.Phone != nil {
		v, err := normalizeOptionalOrgField(p.Phone, 32, ErrPhoneInvalid)
		if err != nil {
			return err
		}
		o.phone = v
	}
	if p.Address != nil {
		v, err := normalizeOptionalOrgField(p.Address, 4096, ErrAddressInvalid)
		if err != nil {
			return err
		}
		o.address = v
	}
	if p.Industry != nil {
		v, err := normalizeOptionalOrgField(p.Industry, 128, ErrIndustryInvalid)
		if err != nil {
			return err
		}
		o.industry = v
	}
	updatedAt := p.UpdatedAt
	o.updatedAt = &updatedAt
	return nil
}

func (o *Organization) ID() OrganizationID       { return o.id }
func (o *Organization) Name() string             { return o.name }
func (o *Organization) Slug() string             { return o.slug }
func (o *Organization) LogoObjectID() *uuid.UUID { return cloneUUIDPtr(o.logoObjectID) }
func (o *Organization) Description() *string     { return cloneStringPtr(o.description) }
func (o *Organization) Website() *string         { return cloneStringPtr(o.website) }
func (o *Organization) Phone() *string           { return cloneStringPtr(o.phone) }
func (o *Organization) Address() *string         { return cloneStringPtr(o.address) }
func (o *Organization) Industry() *string        { return cloneStringPtr(o.industry) }
func (o *Organization) IsActive() bool           { return o.isActive }
func (o *Organization) CreatedAt() time.Time     { return o.createdAt }
func (o *Organization) UpdatedAt() *time.Time    { return cloneTimePtr(o.updatedAt) }

func (o *Organization) PrimaryEmail() *string {
	if o.primaryEmail == nil {
		return nil
	}
	value := o.primaryEmail.String()
	return &value
}

func (o *Organization) Status() OrganizationStatus {
	if o.isActive {
		return OrganizationStatusActive
	}
	return OrganizationStatusArchived
}

func normalizeName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 255 {
		return "", ErrNameInvalid
	}
	return s, nil
}

func normalizeSlug(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 255 {
		return "", ErrSlugInvalid
	}
	return s, nil
}

func normalizeOptionalOrgField(s *string, maxRunes int, invalid error) (*string, error) {
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

func normalizeOptionalOrganizationEmail(s *string) (*Email, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	email, err := NewEmail(v)
	if err != nil {
		return nil, ErrPrimaryEmailInvalid
	}
	return &email, nil
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
