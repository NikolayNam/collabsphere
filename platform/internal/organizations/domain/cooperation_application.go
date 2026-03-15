package domain

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type CooperationApplicationStatus string
type CooperationPriceListStatus string

const (
	CooperationApplicationStatusDraft       CooperationApplicationStatus = "draft"
	CooperationApplicationStatusSubmitted   CooperationApplicationStatus = "submitted"
	CooperationApplicationStatusUnderReview CooperationApplicationStatus = "under_review"
	CooperationApplicationStatusApproved    CooperationApplicationStatus = "approved"
	CooperationApplicationStatusRejected    CooperationApplicationStatus = "rejected"
	CooperationApplicationStatusNeedsInfo   CooperationApplicationStatus = "needs_info"
)

const (
	CooperationPriceListStatusDraft      CooperationPriceListStatus = "draft"
	CooperationPriceListStatusValidating CooperationPriceListStatus = "validating"
	CooperationPriceListStatusVerified   CooperationPriceListStatus = "verified"
	CooperationPriceListStatusPublished  CooperationPriceListStatus = "published"
	CooperationPriceListStatusWithdrawn  CooperationPriceListStatus = "withdrawn"
	CooperationPriceListStatusArchived   CooperationPriceListStatus = "archived"
)

type CooperationApplication struct {
	id                    uuid.UUID
	organizationID        OrganizationID
	status                CooperationApplicationStatus
	confirmationEmail     *Email
	companyName           *string
	representedCategories *string
	minimumOrderAmount    *string
	deliveryGeography     *string
	salesChannels         []string
	storefrontURL         *string
	contactFirstName      *string
	contactLastName       *string
	contactJobTitle       *string
	priceListObjectID     *uuid.UUID
	priceListStatus       CooperationPriceListStatus
	contactEmail          *Email
	contactPhone          *string
	partnerCode           *string
	reviewNote            *string
	reviewerAccountID     *uuid.UUID
	submittedAt           *time.Time
	reviewedAt            *time.Time
	createdAt             time.Time
	updatedAt             *time.Time
}

type NewCooperationApplicationParams struct {
	ID             uuid.UUID
	OrganizationID OrganizationID
	Now            time.Time
}

func NewCooperationApplication(p NewCooperationApplicationParams) (*CooperationApplication, error) {
	if p.ID == uuid.Nil {
		return nil, ErrCooperationApplicationIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}
	updatedAt := p.Now
	return &CooperationApplication{
		id:              p.ID,
		organizationID:  p.OrganizationID,
		status:          CooperationApplicationStatusDraft,
		priceListStatus: CooperationPriceListStatusDraft,
		createdAt:       p.Now,
		updatedAt:       &updatedAt,
	}, nil
}

type RehydrateCooperationApplicationParams struct {
	ID                    uuid.UUID
	OrganizationID        OrganizationID
	Status                string
	ConfirmationEmail     *string
	CompanyName           *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	StorefrontURL         *string
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	PriceListObjectID     *uuid.UUID
	PriceListStatus       string
	ContactEmail          *string
	ContactPhone          *string
	PartnerCode           *string
	ReviewNote            *string
	ReviewerAccountID     *uuid.UUID
	SubmittedAt           *time.Time
	ReviewedAt            *time.Time
	CreatedAt             time.Time
	UpdatedAt             *time.Time
}

func RehydrateCooperationApplication(p RehydrateCooperationApplicationParams) (*CooperationApplication, error) {
	if p.ID == uuid.Nil {
		return nil, ErrCooperationApplicationIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.CreatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt != nil && p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}
	if p.SubmittedAt != nil && p.SubmittedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}
	if p.ReviewedAt != nil && p.ReviewedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	status, err := normalizeCooperationApplicationStatus(p.Status)
	if err != nil {
		return nil, err
	}
	confirmationEmail, err := normalizeOptionalOrganizationEmail(p.ConfirmationEmail)
	if err != nil {
		return nil, ErrCooperationConfirmationEmailInvalid
	}
	companyName, err := normalizeOptionalOrgField(p.CompanyName, 255, ErrCooperationCompanyNameInvalid)
	if err != nil {
		return nil, err
	}
	representedCategories, err := normalizeOptionalOrgField(p.RepresentedCategories, 4096, ErrCooperationRepresentedCategoriesInvalid)
	if err != nil {
		return nil, err
	}
	minimumOrderAmount, err := normalizeOptionalOrgField(p.MinimumOrderAmount, 128, ErrCooperationMinimumOrderAmountInvalid)
	if err != nil {
		return nil, err
	}
	deliveryGeography, err := normalizeOptionalOrgField(p.DeliveryGeography, 4096, ErrCooperationDeliveryGeographyInvalid)
	if err != nil {
		return nil, err
	}
	salesChannels, err := normalizeSalesChannels(p.SalesChannels)
	if err != nil {
		return nil, err
	}
	storefrontURL, err := normalizeOptionalOrgField(p.StorefrontURL, 512, ErrCooperationStorefrontURLInvalid)
	if err != nil {
		return nil, err
	}
	contactFirstName, err := normalizeOptionalOrgField(p.ContactFirstName, 128, ErrCooperationContactFirstNameInvalid)
	if err != nil {
		return nil, err
	}
	contactLastName, err := normalizeOptionalOrgField(p.ContactLastName, 128, ErrCooperationContactLastNameInvalid)
	if err != nil {
		return nil, err
	}
	contactJobTitle, err := normalizeOptionalOrgField(p.ContactJobTitle, 128, ErrCooperationContactJobTitleInvalid)
	if err != nil {
		return nil, err
	}
	contactEmail, err := normalizeOptionalOrganizationEmail(p.ContactEmail)
	if err != nil {
		return nil, ErrCooperationContactEmailInvalid
	}
	contactPhone, err := normalizeOptionalOrgField(p.ContactPhone, 32, ErrCooperationContactPhoneInvalid)
	if err != nil {
		return nil, err
	}
	partnerCode, err := normalizeOptionalOrgField(p.PartnerCode, 128, ErrCooperationPartnerCodeInvalid)
	if err != nil {
		return nil, err
	}
	reviewNote, err := normalizeOptionalOrgField(p.ReviewNote, 4096, ErrCooperationReviewNoteInvalid)
	if err != nil {
		return nil, err
	}
	priceListStatus, err := normalizeCooperationPriceListStatusOrDefault(p.PriceListStatus)
	if err != nil {
		return nil, err
	}

	return &CooperationApplication{
		id:                    p.ID,
		organizationID:        p.OrganizationID,
		status:                status,
		confirmationEmail:     confirmationEmail,
		companyName:           companyName,
		representedCategories: representedCategories,
		minimumOrderAmount:    minimumOrderAmount,
		deliveryGeography:     deliveryGeography,
		salesChannels:         salesChannels,
		storefrontURL:         storefrontURL,
		contactFirstName:      contactFirstName,
		contactLastName:       contactLastName,
		contactJobTitle:       contactJobTitle,
		priceListObjectID:     cloneUUIDPtr(p.PriceListObjectID),
		priceListStatus:       priceListStatus,
		contactEmail:          contactEmail,
		contactPhone:          contactPhone,
		partnerCode:           partnerCode,
		reviewNote:            reviewNote,
		reviewerAccountID:     cloneUUIDPtr(p.ReviewerAccountID),
		submittedAt:           cloneTimePtr(p.SubmittedAt),
		reviewedAt:            cloneTimePtr(p.ReviewedAt),
		createdAt:             p.CreatedAt,
		updatedAt:             cloneTimePtr(p.UpdatedAt),
	}, nil
}

type CooperationApplicationPatch struct {
	ConfirmationEmail     *string
	CompanyName           *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	StorefrontURL         *string
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	PriceListObjectID     *uuid.UUID
	PriceListStatus       *string
	ClearPriceList        bool
	ContactEmail          *string
	ContactPhone          *string
	PartnerCode           *string
	UpdatedAt             time.Time
}

func (a *CooperationApplication) ApplyPatch(p CooperationApplicationPatch) error {
	if a == nil {
		return ErrCooperationApplicationIDEmpty
	}
	if p.UpdatedAt.IsZero() {
		return ErrNowRequired
	}

	if p.ConfirmationEmail != nil {
		v, err := normalizeOptionalOrganizationEmail(p.ConfirmationEmail)
		if err != nil {
			return ErrCooperationConfirmationEmailInvalid
		}
		a.confirmationEmail = v
	}
	if p.CompanyName != nil {
		v, err := normalizeOptionalOrgField(p.CompanyName, 255, ErrCooperationCompanyNameInvalid)
		if err != nil {
			return err
		}
		a.companyName = v
	}
	if p.RepresentedCategories != nil {
		v, err := normalizeOptionalOrgField(p.RepresentedCategories, 4096, ErrCooperationRepresentedCategoriesInvalid)
		if err != nil {
			return err
		}
		a.representedCategories = v
	}
	if p.MinimumOrderAmount != nil {
		v, err := normalizeOptionalOrgField(p.MinimumOrderAmount, 128, ErrCooperationMinimumOrderAmountInvalid)
		if err != nil {
			return err
		}
		a.minimumOrderAmount = v
	}
	if p.DeliveryGeography != nil {
		v, err := normalizeOptionalOrgField(p.DeliveryGeography, 4096, ErrCooperationDeliveryGeographyInvalid)
		if err != nil {
			return err
		}
		a.deliveryGeography = v
	}
	if p.SalesChannels != nil {
		v, err := normalizeSalesChannels(p.SalesChannels)
		if err != nil {
			return err
		}
		a.salesChannels = v
	}
	if p.StorefrontURL != nil {
		v, err := normalizeOptionalOrgField(p.StorefrontURL, 512, ErrCooperationStorefrontURLInvalid)
		if err != nil {
			return err
		}
		a.storefrontURL = v
	}
	if p.ContactFirstName != nil {
		v, err := normalizeOptionalOrgField(p.ContactFirstName, 128, ErrCooperationContactFirstNameInvalid)
		if err != nil {
			return err
		}
		a.contactFirstName = v
	}
	if p.ContactLastName != nil {
		v, err := normalizeOptionalOrgField(p.ContactLastName, 128, ErrCooperationContactLastNameInvalid)
		if err != nil {
			return err
		}
		a.contactLastName = v
	}
	if p.ContactJobTitle != nil {
		v, err := normalizeOptionalOrgField(p.ContactJobTitle, 128, ErrCooperationContactJobTitleInvalid)
		if err != nil {
			return err
		}
		a.contactJobTitle = v
	}
	if p.ClearPriceList {
		a.priceListObjectID = nil
	} else if p.PriceListObjectID != nil {
		a.priceListObjectID = cloneUUIDPtr(p.PriceListObjectID)
	}
	if p.PriceListStatus != nil {
		v, err := normalizeCooperationPriceListStatusOrDefault(*p.PriceListStatus)
		if err != nil {
			return err
		}
		a.priceListStatus = v
	}
	if p.ContactEmail != nil {
		v, err := normalizeOptionalOrganizationEmail(p.ContactEmail)
		if err != nil {
			return ErrCooperationContactEmailInvalid
		}
		a.contactEmail = v
	}
	if p.ContactPhone != nil {
		v, err := normalizeOptionalOrgField(p.ContactPhone, 32, ErrCooperationContactPhoneInvalid)
		if err != nil {
			return err
		}
		a.contactPhone = v
	}
	if p.PartnerCode != nil {
		v, err := normalizeOptionalOrgField(p.PartnerCode, 128, ErrCooperationPartnerCodeInvalid)
		if err != nil {
			return err
		}
		a.partnerCode = v
	}
	updatedAt := p.UpdatedAt
	a.updatedAt = &updatedAt
	if a.status == CooperationApplicationStatusRejected || a.status == CooperationApplicationStatusNeedsInfo {
		a.status = CooperationApplicationStatusDraft
		a.reviewedAt = nil
		a.reviewerAccountID = nil
		a.reviewNote = nil
	}
	return nil
}

func (a *CooperationApplication) MarkSubmitted(now time.Time) error {
	if a == nil {
		return ErrCooperationApplicationIDEmpty
	}
	if now.IsZero() {
		return ErrNowRequired
	}
	if a.confirmationEmail == nil || a.companyName == nil || a.representedCategories == nil || a.minimumOrderAmount == nil || a.deliveryGeography == nil || len(a.salesChannels) == 0 || a.contactFirstName == nil || a.contactLastName == nil || a.contactJobTitle == nil || a.priceListObjectID == nil || a.contactEmail == nil || a.contactPhone == nil {
		return ErrCooperationApplicationIncomplete
	}
	a.status = CooperationApplicationStatusSubmitted
	a.submittedAt = &now
	a.reviewedAt = nil
	a.reviewerAccountID = nil
	a.reviewNote = nil
	a.updatedAt = &now
	return nil
}

func (a *CooperationApplication) ID() uuid.UUID                        { return a.id }
func (a *CooperationApplication) OrganizationID() OrganizationID       { return a.organizationID }
func (a *CooperationApplication) Status() CooperationApplicationStatus { return a.status }
func (a *CooperationApplication) ConfirmationEmail() *string {
	return emailToStringPtr(a.confirmationEmail)
}
func (a *CooperationApplication) CompanyName() *string { return cloneStringPtr(a.companyName) }
func (a *CooperationApplication) RepresentedCategories() *string {
	return cloneStringPtr(a.representedCategories)
}
func (a *CooperationApplication) MinimumOrderAmount() *string {
	return cloneStringPtr(a.minimumOrderAmount)
}
func (a *CooperationApplication) DeliveryGeography() *string {
	return cloneStringPtr(a.deliveryGeography)
}
func (a *CooperationApplication) SalesChannels() []string { return cloneStrings(a.salesChannels) }
func (a *CooperationApplication) StorefrontURL() *string  { return cloneStringPtr(a.storefrontURL) }
func (a *CooperationApplication) ContactFirstName() *string {
	return cloneStringPtr(a.contactFirstName)
}
func (a *CooperationApplication) ContactLastName() *string { return cloneStringPtr(a.contactLastName) }
func (a *CooperationApplication) ContactJobTitle() *string { return cloneStringPtr(a.contactJobTitle) }
func (a *CooperationApplication) PriceListObjectID() *uuid.UUID {
	return cloneUUIDPtr(a.priceListObjectID)
}
func (a *CooperationApplication) PriceListStatus() CooperationPriceListStatus {
	return a.priceListStatus
}
func (a *CooperationApplication) ContactEmail() *string { return emailToStringPtr(a.contactEmail) }
func (a *CooperationApplication) ContactPhone() *string { return cloneStringPtr(a.contactPhone) }
func (a *CooperationApplication) PartnerCode() *string  { return cloneStringPtr(a.partnerCode) }
func (a *CooperationApplication) ReviewNote() *string   { return cloneStringPtr(a.reviewNote) }
func (a *CooperationApplication) ReviewerAccountID() *uuid.UUID {
	return cloneUUIDPtr(a.reviewerAccountID)
}
func (a *CooperationApplication) SubmittedAt() *time.Time { return cloneTimePtr(a.submittedAt) }
func (a *CooperationApplication) ReviewedAt() *time.Time  { return cloneTimePtr(a.reviewedAt) }
func (a *CooperationApplication) CreatedAt() time.Time    { return a.createdAt }
func (a *CooperationApplication) UpdatedAt() *time.Time   { return cloneTimePtr(a.updatedAt) }

func normalizeCooperationApplicationStatus(value string) (CooperationApplicationStatus, error) {
	switch CooperationApplicationStatus(strings.TrimSpace(value)) {
	case CooperationApplicationStatusDraft,
		CooperationApplicationStatusSubmitted,
		CooperationApplicationStatusUnderReview,
		CooperationApplicationStatusApproved,
		CooperationApplicationStatusRejected,
		CooperationApplicationStatusNeedsInfo:
		return CooperationApplicationStatus(strings.TrimSpace(value)), nil
	default:
		return "", ErrCooperationApplicationStatusInvalid
	}
}

func normalizeCooperationPriceListStatusOrDefault(value string) (CooperationPriceListStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return CooperationPriceListStatusDraft, nil
	}
	switch CooperationPriceListStatus(normalized) {
	case CooperationPriceListStatusDraft,
		CooperationPriceListStatusValidating,
		CooperationPriceListStatusVerified,
		CooperationPriceListStatusPublished,
		CooperationPriceListStatusWithdrawn,
		CooperationPriceListStatusArchived:
		return CooperationPriceListStatus(normalized), nil
	default:
		return "", ErrCooperationPriceListStatusInvalid
	}
}

func normalizeSalesChannels(values []string) ([]string, error) {
	if values == nil {
		return nil, nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if utf8.RuneCountInString(trimmed) > 64 {
			return nil, ErrCooperationSalesChannelsInvalid
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) > 12 {
		return nil, ErrCooperationSalesChannelsInvalid
	}
	return out, nil
}

func emailToStringPtr(email *Email) *string {
	if email == nil {
		return nil
	}
	value := email.String()
	return &value
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func MarshalSalesChannels(values []string) ([]byte, error) {
	normalized, err := normalizeSalesChannels(values)
	if err != nil {
		return nil, err
	}
	if normalized == nil {
		normalized = []string{}
	}
	return json.Marshal(normalized)
}

func UnmarshalSalesChannels(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return normalizeSalesChannels(out)
}
