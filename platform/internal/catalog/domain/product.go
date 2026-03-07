package domain

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

var moneyPattern = regexp.MustCompile(`^\d{1,12}(\.\d{1,2})?$`)
var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Product struct {
	id             ProductID
	organizationID orgdomain.OrganizationID
	categoryID     *ProductCategoryID
	name           string
	description    *string
	sku            *string
	priceAmount    *string
	currencyCode   *string
	isActive       bool
	createdAt      time.Time
	updatedAt      *time.Time
}

type NewProductParams struct {
	ID             ProductID
	OrganizationID orgdomain.OrganizationID
	CategoryID     *ProductCategoryID
	Name           string
	Description    *string
	SKU            *string
	PriceAmount    *string
	CurrencyCode   *string
	IsActive       *bool
	Now            time.Time
}

type RehydrateProductParams struct {
	ID             ProductID
	OrganizationID orgdomain.OrganizationID
	CategoryID     *ProductCategoryID
	Name           string
	Description    *string
	SKU            *string
	PriceAmount    *string
	CurrencyCode   *string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewProduct(p NewProductParams) (*Product, error) {
	if p.ID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrProductIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	name, err := normalizeProductName(p.Name)
	if err != nil {
		return nil, err
	}
	description := normalizeOptionalText(p.Description)
	sku, err := normalizeSKU(p.SKU)
	if err != nil {
		return nil, err
	}
	priceAmount, currencyCode, err := normalizeMoney(p.PriceAmount, p.CurrencyCode)
	if err != nil {
		return nil, err
	}
	isActive := true
	if p.IsActive != nil {
		isActive = *p.IsActive
	}

	updatedAt := p.Now
	return &Product{
		id:             p.ID,
		organizationID: p.OrganizationID,
		categoryID:     cloneProductCategoryIDPtr(p.CategoryID),
		name:           name,
		description:    description,
		sku:            sku,
		priceAmount:    priceAmount,
		currencyCode:   currencyCode,
		isActive:       isActive,
		createdAt:      p.Now,
		updatedAt:      &updatedAt,
	}, nil
}

func RehydrateProduct(p RehydrateProductParams) (*Product, error) {
	if p.ID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrProductIDEmpty
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	name, err := normalizeProductName(p.Name)
	if err != nil {
		return nil, err
	}
	description := normalizeOptionalText(p.Description)
	sku, err := normalizeSKU(p.SKU)
	if err != nil {
		return nil, err
	}
	priceAmount, currencyCode, err := normalizeMoney(p.PriceAmount, p.CurrencyCode)
	if err != nil {
		return nil, err
	}

	updatedAt := p.UpdatedAt
	return &Product{
		id:             p.ID,
		organizationID: p.OrganizationID,
		categoryID:     cloneProductCategoryIDPtr(p.CategoryID),
		name:           name,
		description:    description,
		sku:            sku,
		priceAmount:    priceAmount,
		currencyCode:   currencyCode,
		isActive:       p.IsActive,
		createdAt:      p.CreatedAt,
		updatedAt:      &updatedAt,
	}, nil
}

func (p *Product) ID() ProductID                            { return p.id }
func (p *Product) OrganizationID() orgdomain.OrganizationID { return p.organizationID }
func (p *Product) CategoryID() *ProductCategoryID           { return cloneProductCategoryIDPtr(p.categoryID) }
func (p *Product) Name() string                             { return p.name }
func (p *Product) Description() *string                     { return cloneStringPtr(p.description) }
func (p *Product) SKU() *string                             { return cloneStringPtr(p.sku) }
func (p *Product) PriceAmount() *string                     { return cloneStringPtr(p.priceAmount) }
func (p *Product) CurrencyCode() *string                    { return cloneStringPtr(p.currencyCode) }
func (p *Product) IsActive() bool                           { return p.isActive }
func (p *Product) CreatedAt() time.Time                     { return p.createdAt }
func (p *Product) UpdatedAt() *time.Time                    { return cloneTimePtr(p.updatedAt) }

func normalizeProductName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 255 {
		return "", ErrProductNameInvalid
	}
	return s, nil
}

func normalizeOptionalText(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeSKU(s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > 128 {
		return nil, ErrProductSKUInvalid
	}
	return &trimmed, nil
}

func normalizeMoney(price *string, currency *string) (*string, *string, error) {
	priceNorm := normalizeOptionalText(price)
	currencyNorm := normalizeOptionalText(currency)
	if priceNorm == nil && currencyNorm == nil {
		return nil, nil, nil
	}
	if priceNorm == nil || currencyNorm == nil {
		return nil, nil, ErrProductPriceInvalid
	}
	if !moneyPattern.MatchString(*priceNorm) {
		return nil, nil, ErrProductPriceInvalid
	}
	value, err := strconv.ParseFloat(*priceNorm, 64)
	if err != nil || value < 0 {
		return nil, nil, ErrProductPriceInvalid
	}
	upper := strings.ToUpper(*currencyNorm)
	if !currencyPattern.MatchString(upper) {
		return nil, nil, ErrProductCurrencyInvalid
	}
	return priceNorm, &upper, nil
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}
