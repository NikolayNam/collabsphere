package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCooperationApplicationMarkSubmittedRequiresRequiredFields(t *testing.T) {
	organizationID := NewOrganizationID()
	application, err := NewCooperationApplication(NewCooperationApplicationParams{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Now:            time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewCooperationApplication: %v", err)
	}
	if err := application.MarkSubmitted(time.Date(2026, 3, 8, 13, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expected incomplete cooperation application error")
	}

	priceListID := uuid.New()
	if err := application.ApplyPatch(CooperationApplicationPatch{
		ConfirmationEmail:     ptrStringOnboarding("confirm@example.com"),
		CompanyName:           ptrStringOnboarding("Acme Supply"),
		RepresentedCategories: ptrStringOnboarding("Beverages, snacks"),
		MinimumOrderAmount:    ptrStringOnboarding("100000 RUB"),
		DeliveryGeography:     ptrStringOnboarding("Moscow and region"),
		SalesChannels:         []string{"Retail", "HoReCa"},
		ContactFirstName:      ptrStringOnboarding("Ivan"),
		ContactLastName:       ptrStringOnboarding("Petrov"),
		ContactJobTitle:       ptrStringOnboarding("Sales Manager"),
		PriceListObjectID:     &priceListID,
		ContactEmail:          ptrStringOnboarding("sales@example.com"),
		ContactPhone:          ptrStringOnboarding("+74951234567"),
		UpdatedAt:             time.Date(2026, 3, 8, 14, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("ApplyPatch: %v", err)
	}
	if err := application.MarkSubmitted(time.Date(2026, 3, 8, 15, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("MarkSubmitted: %v", err)
	}
	if application.Status() != CooperationApplicationStatusSubmitted {
		t.Fatalf("unexpected status: %s", application.Status())
	}
	if application.SubmittedAt() == nil {
		t.Fatal("submittedAt was not set")
	}
}

func TestNewOrganizationLegalDocument(t *testing.T) {
	organizationID := NewOrganizationID()
	uploadedBy := uuid.New()
	document, err := NewOrganizationLegalDocument(NewOrganizationLegalDocumentParams{
		ID:                  uuid.New(),
		OrganizationID:      organizationID,
		DocumentType:        "registration_certificate",
		ObjectID:            uuid.New(),
		Title:               "Company registration extract",
		UploadedByAccountID: &uploadedBy,
		Now:                 time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewOrganizationLegalDocument: %v", err)
	}
	if document.Status() != OrganizationLegalDocumentStatusPending {
		t.Fatalf("unexpected status: %s", document.Status())
	}
	if got := document.UploadedByAccountID(); got == nil || *got != uploadedBy {
		t.Fatalf("unexpected uploadedByAccountID: %v", got)
	}
}

func ptrStringOnboarding(value string) *string {
	return &value
}
