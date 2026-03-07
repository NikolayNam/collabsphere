package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Group struct {
	id          GroupID
	name        string
	slug        string
	description *string
	isActive    bool
	createdAt   time.Time
	updatedAt   *time.Time
}

type NewGroupParams struct {
	ID          GroupID
	Name        string
	Slug        string
	Description *string
	Now         time.Time
}

type RehydrateGroupParams struct {
	ID          GroupID
	Name        string
	Slug        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewGroup(p NewGroupParams) (*Group, error) {
	if p.ID.IsZero() {
		return nil, ErrGroupIDEmpty
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
	description, err := normalizeDescription(p.Description)
	if err != nil {
		return nil, err
	}

	updatedAt := p.Now

	return &Group{
		id:          p.ID,
		name:        name,
		slug:        slug,
		description: description,
		isActive:    true,
		createdAt:   p.Now,
		updatedAt:   &updatedAt,
	}, nil
}

func RehydrateGroup(p RehydrateGroupParams) (*Group, error) {
	if p.ID.IsZero() {
		return nil, ErrGroupIDEmpty
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
	description, err := normalizeDescription(p.Description)
	if err != nil {
		return nil, err
	}

	updatedAt := p.UpdatedAt

	return &Group{
		id:          p.ID,
		name:        name,
		slug:        slug,
		description: description,
		isActive:    p.IsActive,
		createdAt:   p.CreatedAt,
		updatedAt:   &updatedAt,
	}, nil
}

func (g *Group) ID() GroupID           { return g.id }
func (g *Group) Name() string          { return g.name }
func (g *Group) Slug() string          { return g.slug }
func (g *Group) IsActive() bool        { return g.isActive }
func (g *Group) CreatedAt() time.Time  { return g.createdAt }
func (g *Group) UpdatedAt() *time.Time { return cloneTimePtr(g.updatedAt) }
func (g *Group) Description() *string  { return cloneStringPtr(g.description) }

func normalizeName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 255 {
		return "", ErrGroupNameInvalid
	}
	return value, nil
}

func normalizeSlug(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 255 {
		return "", ErrGroupSlugInvalid
	}
	return value, nil
}

func normalizeDescription(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(normalized) > 2000 {
		return nil, ErrGroupDescriptionInvalid
	}
	return &normalized, nil
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
