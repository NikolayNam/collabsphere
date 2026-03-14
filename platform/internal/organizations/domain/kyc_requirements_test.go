package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildOrganizationKYCRequirementsCurrentlyDue(t *testing.T) {
	result := BuildOrganizationKYCRequirements(OrganizationKYCRequirementsInput{
		OrganizationID: uuid.New(),
		Now:            time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})

	if result == nil {
		t.Fatal("result = nil")
	}
	if result.Status != OrganizationKYCRequirementsStatusCurrentlyDue {
		t.Fatalf("status = %q, want currently_due", result.Status)
	}
	if len(result.CurrentlyDue) != 2 {
		t.Fatalf("currentlyDue len = %d, want 2", len(result.CurrentlyDue))
	}
}

func TestBuildOrganizationKYCRequirementsPendingVerification(t *testing.T) {
	documentID := uuid.New()
	result := BuildOrganizationKYCRequirements(OrganizationKYCRequirementsInput{
		OrganizationID: uuid.New(),
		CooperationApplication: &OrganizationKYCCooperationApplicationInput{
			Status:                "submitted",
			ConfirmationEmail:     strPtrKYC("confirm@example.com"),
			CompanyName:           strPtrKYC("Acme"),
			RepresentedCategories: strPtrKYC("Food"),
			MinimumOrderAmount:    strPtrKYC("1000"),
			DeliveryGeography:     strPtrKYC("Moscow"),
			SalesChannels:         []string{"Retail"},
			PriceListObjectID:     uuidPtrKYC(uuid.New()),
			ContactFirstName:      strPtrKYC("Ivan"),
			ContactLastName:       strPtrKYC("Petrov"),
			ContactJobTitle:       strPtrKYC("Manager"),
			ContactEmail:          strPtrKYC("sales@example.com"),
			ContactPhone:          strPtrKYC("+79990000000"),
		},
		LegalDocuments: []OrganizationKYCLegalDocumentInput{{
			ID:           documentID,
			DocumentType: "inn_certificate",
			Status:       "pending",
			Verification: &OrganizationLegalDocumentVerification{
				DocumentID:     documentID,
				Verdict:        OrganizationLegalDocumentVerificationVerdictApproved,
				Summary:        "Machine verification passed.",
				RequiredFields: []string{"inn"},
			},
		}},
		Now: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != OrganizationKYCRequirementsStatusPendingVerification {
		t.Fatalf("status = %q, want pending_verification", result.Status)
	}
	if len(result.PendingVerification) != 2 {
		t.Fatalf("pendingVerification len = %d, want 2", len(result.PendingVerification))
	}
}

func TestBuildOrganizationKYCRequirementsNeedsInfo(t *testing.T) {
	documentID := uuid.New()
	result := BuildOrganizationKYCRequirements(OrganizationKYCRequirementsInput{
		OrganizationID: uuid.New(),
		CooperationApplication: &OrganizationKYCCooperationApplicationInput{
			Status:     "approved",
			ReviewNote: strPtrKYC("Approved"),
		},
		LegalDocuments: []OrganizationKYCLegalDocumentInput{{
			ID:           documentID,
			DocumentType: "inn_certificate",
			Status:       "rejected",
			ReviewNote:   strPtrKYC("Upload clearer scan"),
		}},
		Now: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != OrganizationKYCRequirementsStatusNeedsInfo {
		t.Fatalf("status = %q, want needs_info", result.Status)
	}
	if result.DisabledReason == nil || *result.DisabledReason != "requirements_past_due" {
		t.Fatalf("disabledReason = %v, want requirements_past_due", result.DisabledReason)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors len = %d, want 1", len(result.Errors))
	}
}

func TestBuildOrganizationKYCRequirementsVerified(t *testing.T) {
	documentID := uuid.New()
	result := BuildOrganizationKYCRequirements(OrganizationKYCRequirementsInput{
		OrganizationID: uuid.New(),
		CooperationApplication: &OrganizationKYCCooperationApplicationInput{
			Status:                "approved",
			ConfirmationEmail:     strPtrKYC("confirm@example.com"),
			CompanyName:           strPtrKYC("Acme"),
			RepresentedCategories: strPtrKYC("Food"),
			MinimumOrderAmount:    strPtrKYC("1000"),
			DeliveryGeography:     strPtrKYC("Moscow"),
			SalesChannels:         []string{"Retail"},
			PriceListObjectID:     uuidPtrKYC(uuid.New()),
			ContactFirstName:      strPtrKYC("Ivan"),
			ContactLastName:       strPtrKYC("Petrov"),
			ContactJobTitle:       strPtrKYC("Manager"),
			ContactEmail:          strPtrKYC("sales@example.com"),
			ContactPhone:          strPtrKYC("+79990000000"),
		},
		LegalDocuments: []OrganizationKYCLegalDocumentInput{{
			ID:           documentID,
			DocumentType: "inn_certificate",
			Status:       "approved",
		}},
		Now: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
	})

	if result.Status != OrganizationKYCRequirementsStatusVerified {
		t.Fatalf("status = %q, want verified", result.Status)
	}
	if len(result.CurrentlyDue)+len(result.PendingVerification)+len(result.Errors) != 0 {
		t.Fatalf("unexpected outstanding items: currentlyDue=%d pending=%d errors=%d", len(result.CurrentlyDue), len(result.PendingVerification), len(result.Errors))
	}
}

func strPtrKYC(value string) *string {
	return &value
}

func uuidPtrKYC(value uuid.UUID) *uuid.UUID {
	return &value
}
