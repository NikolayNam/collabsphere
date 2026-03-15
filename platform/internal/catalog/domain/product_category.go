package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type ProductCategory struct {
	id             ProductCategoryID
	organizationID orgdomain.OrganizationID
	parentID       *ProductCategoryID
	templateID     *uuid.UUID
	status         ProductCategoryStatus
	code           string
	name           string
	sortOrder      int64
	createdAt      time.Time
	updatedAt      *time.Time
}

type NewProductCategoryParams struct {
	ID             ProductCategoryID
	OrganizationID orgdomain.OrganizationID
	ParentID       *ProductCategoryID
	Status         string
	Code           string
	Name           string
	SortOrder      int64
	Now            time.Time
}

type RehydrateProductCategoryParams struct {
	ID             ProductCategoryID
	OrganizationID orgdomain.OrganizationID
	ParentID       *ProductCategoryID
	TemplateID     *uuid.UUID
	Status         string
	Code           string
	Name           string
	SortOrder      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewProductCategory(p NewProductCategoryParams) (*ProductCategory, error) {
	if p.ID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrProductCategoryIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	code, err := normalizeCategoryCode(p.Code)
	if err != nil {
		return nil, err
	}
	name, err := normalizeCategoryName(p.Name)
	if err != nil {
		return nil, err
	}
	status, err := normalizeCategoryStatusOrDefault(p.Status)
	if err != nil {
		return nil, err
	}
	if p.SortOrder < 0 {
		return nil, ErrProductCategorySortInvalid
	}

	updatedAt := p.Now
	return &ProductCategory{
		id:             p.ID,
		organizationID: p.OrganizationID,
		parentID:       cloneProductCategoryIDPtr(p.ParentID),
		status:         status,
		code:           code,
		name:           name,
		sortOrder:      p.SortOrder,
		createdAt:      p.Now,
		updatedAt:      &updatedAt,
	}, nil
}

func RehydrateProductCategory(p RehydrateProductCategoryParams) (*ProductCategory, error) {
	if p.ID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrProductCategoryIDEmpty
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	code, err := normalizeCategoryCode(p.Code)
	if err != nil {
		return nil, err
	}
	name, err := normalizeCategoryName(p.Name)
	if err != nil {
		return nil, err
	}
	status, err := normalizeCategoryStatusOrDefault(p.Status)
	if err != nil {
		return nil, err
	}
	if p.SortOrder < 0 {
		return nil, ErrProductCategorySortInvalid
	}

	updatedAt := p.UpdatedAt
	return &ProductCategory{
		id:             p.ID,
		organizationID: p.OrganizationID,
		parentID:       cloneProductCategoryIDPtr(p.ParentID),
		templateID:     cloneUUIDPtr(p.TemplateID),
		status:         status,
		code:           code,
		name:           name,
		sortOrder:      p.SortOrder,
		createdAt:      p.CreatedAt,
		updatedAt:      &updatedAt,
	}, nil
}

func (c *ProductCategory) ID() ProductCategoryID                    { return c.id }
func (c *ProductCategory) OrganizationID() orgdomain.OrganizationID { return c.organizationID }
func (c *ProductCategory) ParentID() *ProductCategoryID             { return cloneProductCategoryIDPtr(c.parentID) }
func (c *ProductCategory) TemplateID() *uuid.UUID                   { return cloneUUIDPtr(c.templateID) }
func (c *ProductCategory) Status() ProductCategoryStatus            { return c.status }
func (c *ProductCategory) Code() string                             { return c.code }
func (c *ProductCategory) Name() string                             { return c.name }
func (c *ProductCategory) SortOrder() int64                         { return c.sortOrder }
func (c *ProductCategory) CreatedAt() time.Time                     { return c.createdAt }
func (c *ProductCategory) UpdatedAt() *time.Time                    { return cloneTimePtr(c.updatedAt) }

func normalizeCategoryCode(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 128 {
		return "", ErrProductCategoryCodeInvalid
	}
	return s, nil
}

func normalizeCategoryName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 255 {
		return "", ErrProductCategoryNameInvalid
	}
	return s, nil
}

func cloneProductCategoryIDPtr(id *ProductCategoryID) *ProductCategoryID {
	if id == nil {
		return nil
	}
	v := *id
	return &v
}

func cloneUUIDPtr(id *uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	v := *id
	return &v
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}
