package domain

import (
    "strings"
    "time"
    "unicode/utf8"
)

type Organization struct {
    id        OrganizationID
    name      string
    slug      string
    isActive  bool
    createdAt time.Time
    updatedAt *time.Time
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
    ID        OrganizationID
    Name      string
    Slug      string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
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

    updatedAt := p.UpdatedAt

    return &Organization{
        id:        p.ID,
        name:      name,
        slug:      slug,
        isActive:  p.IsActive,
        createdAt: p.CreatedAt,
        updatedAt: &updatedAt,
    }, nil
}

func (o *Organization) ID() OrganizationID { return o.id }
func (o *Organization) Name() string        { return o.name }
func (o *Organization) Slug() string        { return o.slug }
func (o *Organization) IsActive() bool      { return o.isActive }
func (o *Organization) CreatedAt() time.Time { return o.createdAt }
func (o *Organization) UpdatedAt() *time.Time { return cloneTimePtr(o.updatedAt) }

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

func cloneTimePtr(t *time.Time) *time.Time {
    if t == nil {
        return nil
    }
    v := *t
    return &v
}
