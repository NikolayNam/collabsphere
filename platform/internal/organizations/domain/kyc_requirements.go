package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrganizationKYCRequirementsStatus string

const (
	OrganizationKYCRequirementsStatusCurrentlyDue        OrganizationKYCRequirementsStatus = "currently_due"
	OrganizationKYCRequirementsStatusPendingVerification OrganizationKYCRequirementsStatus = "pending_verification"
	OrganizationKYCRequirementsStatusNeedsInfo           OrganizationKYCRequirementsStatus = "needs_info"
	OrganizationKYCRequirementsStatusVerified            OrganizationKYCRequirementsStatus = "verified"
)

type OrganizationKYCRequirementCategory string

const (
	OrganizationKYCRequirementCategoryCooperationField  OrganizationKYCRequirementCategory = "cooperation_field"
	OrganizationKYCRequirementCategoryCooperationReview OrganizationKYCRequirementCategory = "cooperation_review"
	OrganizationKYCRequirementCategoryLegalDocument     OrganizationKYCRequirementCategory = "legal_document"
	OrganizationKYCRequirementCategoryLegalReview       OrganizationKYCRequirementCategory = "legal_document_review"
	OrganizationKYCRequirementCategoryMachineCheck      OrganizationKYCRequirementCategory = "machine_check"
)

type OrganizationKYCRequirementItem struct {
	Code         string
	Category     OrganizationKYCRequirementCategory
	Title        string
	Description  string
	Field        *string
	DocumentID   *uuid.UUID
	DocumentType *string
	Reason       *string
}

type OrganizationKYCRequirements struct {
	OrganizationID      uuid.UUID
	Status              OrganizationKYCRequirementsStatus
	DisabledReason      *string
	CurrentlyDue        []OrganizationKYCRequirementItem
	PendingVerification []OrganizationKYCRequirementItem
	EventuallyDue       []OrganizationKYCRequirementItem
	Errors              []OrganizationKYCRequirementItem
	CheckedAt           time.Time
}

type OrganizationKYCRequirementsInput struct {
	OrganizationID         uuid.UUID
	CooperationApplication *OrganizationKYCCooperationApplicationInput
	LegalDocuments         []OrganizationKYCLegalDocumentInput
	Now                    time.Time
}

type OrganizationKYCCooperationApplicationInput struct {
	Status                string
	ReviewNote            *string
	ConfirmationEmail     *string
	CompanyName           *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	PriceListObjectID     *uuid.UUID
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	ContactEmail          *string
	ContactPhone          *string
}

type OrganizationKYCLegalDocumentInput struct {
	ID           uuid.UUID
	DocumentType string
	Status       string
	ReviewNote   *string
	Verification *OrganizationLegalDocumentVerification
}

type kycCooperationFieldRule struct {
	code        string
	title       string
	description string
	field       string
	missing     func(*OrganizationKYCCooperationApplicationInput) bool
}

var organizationKYCCooperationFieldRules = []kycCooperationFieldRule{
	{code: "cooperation.confirmation_email", title: "Confirmation email", description: "Provide the organization confirmation email in the cooperation application.", field: "confirmationEmail", missing: func(app *OrganizationKYCCooperationApplicationInput) bool {
		return isBlankString(app.ConfirmationEmail)
	}},
	{code: "cooperation.company_name", title: "Company name", description: "Provide the legal company name in the cooperation application.", field: "companyName", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.CompanyName) }},
	{code: "cooperation.represented_categories", title: "Represented categories", description: "Describe which product categories the organization represents.", field: "representedCategories", missing: func(app *OrganizationKYCCooperationApplicationInput) bool {
		return isBlankString(app.RepresentedCategories)
	}},
	{code: "cooperation.minimum_order_amount", title: "Minimum order amount", description: "Specify the minimum order amount for cooperation.", field: "minimumOrderAmount", missing: func(app *OrganizationKYCCooperationApplicationInput) bool {
		return isBlankString(app.MinimumOrderAmount)
	}},
	{code: "cooperation.delivery_geography", title: "Delivery geography", description: "Specify the delivery geography covered by the organization.", field: "deliveryGeography", missing: func(app *OrganizationKYCCooperationApplicationInput) bool {
		return isBlankString(app.DeliveryGeography)
	}},
	{code: "cooperation.sales_channels", title: "Sales channels", description: "Select at least one sales channel for the cooperation application.", field: "salesChannels", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return len(app.SalesChannels) == 0 }},
	{code: "cooperation.price_list", title: "Price list", description: "Upload and attach a cooperation price list.", field: "priceListObjectId", missing: func(app *OrganizationKYCCooperationApplicationInput) bool {
		return app.PriceListObjectID == nil || *app.PriceListObjectID == uuid.Nil
	}},
	{code: "cooperation.contact_first_name", title: "Contact first name", description: "Provide the first name of the main cooperation contact.", field: "contactFirstName", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.ContactFirstName) }},
	{code: "cooperation.contact_last_name", title: "Contact last name", description: "Provide the last name of the main cooperation contact.", field: "contactLastName", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.ContactLastName) }},
	{code: "cooperation.contact_job_title", title: "Contact job title", description: "Provide the job title of the main cooperation contact.", field: "contactJobTitle", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.ContactJobTitle) }},
	{code: "cooperation.contact_email", title: "Contact email", description: "Provide the email address of the main cooperation contact.", field: "contactEmail", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.ContactEmail) }},
	{code: "cooperation.contact_phone", title: "Contact phone", description: "Provide the phone number of the main cooperation contact.", field: "contactPhone", missing: func(app *OrganizationKYCCooperationApplicationInput) bool { return isBlankString(app.ContactPhone) }},
}

func BuildOrganizationKYCRequirements(input OrganizationKYCRequirementsInput) *OrganizationKYCRequirements {
	if input.OrganizationID == uuid.Nil {
		return nil
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	result := &OrganizationKYCRequirements{
		OrganizationID:      input.OrganizationID,
		CheckedAt:           now,
		CurrentlyDue:        make([]OrganizationKYCRequirementItem, 0, 8),
		PendingVerification: make([]OrganizationKYCRequirementItem, 0, 8),
		EventuallyDue:       make([]OrganizationKYCRequirementItem, 0),
		Errors:              make([]OrganizationKYCRequirementItem, 0, 8),
	}

	appendCooperationKYCRequirements(result, input.CooperationApplication)
	appendLegalDocumentKYCRequirements(result, input.LegalDocuments)

	switch {
	case len(result.Errors) > 0:
		result.Status = OrganizationKYCRequirementsStatusNeedsInfo
		result.DisabledReason = kycStringPtr("requirements_past_due")
	case len(result.CurrentlyDue) > 0:
		result.Status = OrganizationKYCRequirementsStatusCurrentlyDue
	case len(result.PendingVerification) > 0:
		result.Status = OrganizationKYCRequirementsStatusPendingVerification
	default:
		result.Status = OrganizationKYCRequirementsStatusVerified
	}
	return result
}

func appendCooperationKYCRequirements(result *OrganizationKYCRequirements, app *OrganizationKYCCooperationApplicationInput) {
	if result == nil {
		return
	}
	if app == nil {
		result.CurrentlyDue = append(result.CurrentlyDue, OrganizationKYCRequirementItem{
			Code:        "cooperation.application",
			Category:    OrganizationKYCRequirementCategoryCooperationField,
			Title:       "Cooperation application",
			Description: "Create and fill the cooperation application before KYC review can begin.",
		})
		return
	}

	for _, rule := range organizationKYCCooperationFieldRules {
		if !rule.missing(app) {
			continue
		}
		result.CurrentlyDue = append(result.CurrentlyDue, OrganizationKYCRequirementItem{
			Code:        rule.code,
			Category:    OrganizationKYCRequirementCategoryCooperationField,
			Title:       rule.title,
			Description: rule.description,
			Field:       kycStringPtr(rule.field),
		})
	}

	status := CooperationApplicationStatus(strings.TrimSpace(app.Status))
	switch status {
	case CooperationApplicationStatusSubmitted, CooperationApplicationStatusUnderReview:
		result.PendingVerification = append(result.PendingVerification, OrganizationKYCRequirementItem{
			Code:        "cooperation.review",
			Category:    OrganizationKYCRequirementCategoryCooperationReview,
			Title:       "Cooperation application review",
			Description: "The cooperation application is waiting for platform review.",
		})
	case CooperationApplicationStatusRejected, CooperationApplicationStatusNeedsInfo:
		reason := nonEmptyString(app.ReviewNote, "The cooperation application requires changes before KYC can continue.")
		result.Errors = append(result.Errors, OrganizationKYCRequirementItem{
			Code:        "cooperation.review",
			Category:    OrganizationKYCRequirementCategoryCooperationReview,
			Title:       "Cooperation application review",
			Description: "The cooperation application was returned for changes.",
			Reason:      kycStringPtr(reason),
		})
	case CooperationApplicationStatusDraft:
		if len(result.CurrentlyDue) == 0 {
			result.CurrentlyDue = append(result.CurrentlyDue, OrganizationKYCRequirementItem{
				Code:        "cooperation.submit",
				Category:    OrganizationKYCRequirementCategoryCooperationField,
				Title:       "Submit cooperation application",
				Description: "Submit the completed cooperation application to start KYC review.",
			})
		}
	}
}

func appendLegalDocumentKYCRequirements(result *OrganizationKYCRequirements, docs []OrganizationKYCLegalDocumentInput) {
	if result == nil {
		return
	}
	if len(docs) == 0 {
		result.CurrentlyDue = append(result.CurrentlyDue, OrganizationKYCRequirementItem{
			Code:        "legal_document.upload",
			Category:    OrganizationKYCRequirementCategoryLegalDocument,
			Title:       "Legal documents",
			Description: "Upload at least one legal document for organization verification.",
		})
		return
	}

	if approvedDocumentExists(docs) {
		return
	}

	for _, doc := range docs {
		if doc.ID == uuid.Nil {
			continue
		}
		documentID := doc.ID
		documentType := strings.TrimSpace(doc.DocumentType)
		status := OrganizationLegalDocumentStatus(strings.TrimSpace(doc.Status))

		if status == OrganizationLegalDocumentStatusRejected {
			reason := nonEmptyString(doc.ReviewNote, "The legal document was rejected and requires replacement or fixes.")
			result.Errors = append(result.Errors, OrganizationKYCRequirementItem{
				Code:         "legal_document.review",
				Category:     OrganizationKYCRequirementCategoryLegalReview,
				Title:        "Legal document review",
				Description:  "A legal document was rejected during platform review.",
				DocumentID:   &documentID,
				DocumentType: kycStringPtr(documentType),
				Reason:       kycStringPtr(reason),
			})
			return
		}

		if doc.Verification != nil && doc.Verification.Verdict == OrganizationLegalDocumentVerificationVerdictRejected {
			reason := nonEmptyString(&doc.Verification.Summary, "Machine verification rejected the legal document.")
			result.Errors = append(result.Errors, OrganizationKYCRequirementItem{
				Code:         "legal_document.machine_rejected",
				Category:     OrganizationKYCRequirementCategoryMachineCheck,
				Title:        "Machine document verification",
				Description:  "Machine verification found a blocking problem in the uploaded legal document.",
				DocumentID:   &documentID,
				DocumentType: kycStringPtr(documentType),
				Reason:       kycStringPtr(reason),
			})
			return
		}

		description := "The legal document is waiting for platform review."
		if doc.Verification != nil {
			switch doc.Verification.Verdict {
			case OrganizationLegalDocumentVerificationVerdictManualReview:
				description = "The legal document needs manual review before KYC can proceed."
			case OrganizationLegalDocumentVerificationVerdictApproved:
				description = "The legal document passed machine verification and is waiting for platform approval."
			}
		}
		result.PendingVerification = append(result.PendingVerification, OrganizationKYCRequirementItem{
			Code:         "legal_document.review",
			Category:     OrganizationKYCRequirementCategoryLegalReview,
			Title:        "Legal document review",
			Description:  description,
			DocumentID:   &documentID,
			DocumentType: kycStringPtr(documentType),
		})
	}
}

func approvedDocumentExists(docs []OrganizationKYCLegalDocumentInput) bool {
	for _, doc := range docs {
		if OrganizationLegalDocumentStatus(strings.TrimSpace(doc.Status)) == OrganizationLegalDocumentStatusApproved {
			return true
		}
	}
	return false
}

func isBlankString(value *string) bool {
	return strings.TrimSpace(derefVerificationString(value)) == ""
}

func nonEmptyString(value *string, fallback string) string {
	trimmed := strings.TrimSpace(derefVerificationString(value))
	if trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}

func kycStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
