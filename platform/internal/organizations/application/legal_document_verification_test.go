package application

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

func TestBuildOrganizationLegalDocumentVerificationApproved(t *testing.T) {
	document := mustVerificationDocument(t, "inn_certificate")
	analysis := mustVerificationAnalysis(t, document.OrganizationID(), document.ID(), domain.OrganizationLegalDocumentAnalysisStatusCompleted, json.RawMessage(`{"inn":"7701234567","companyName":"Acme","registrationDate":"2022-01-01"}`), stringPtr("inn_certificate"), floatPtr(0.98), nil)

	result := domain.BuildOrganizationLegalDocumentVerification(domain.OrganizationLegalDocumentVerificationInput{
		Document: domain.OrganizationLegalDocumentVerificationDocumentInput{
			ID:             document.ID(),
			OrganizationID: document.OrganizationID().UUID(),
			DocumentType:   document.DocumentType(),
			DocumentStatus: string(document.Status()),
		},
		Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(analysis),
		Now:      time.Now().UTC(),
	})

	if result == nil {
		t.Fatal("result = nil")
	}
	if result.Verdict != domain.OrganizationLegalDocumentVerificationVerdictApproved {
		t.Fatalf("verdict = %q, want approved", result.Verdict)
	}
	if len(result.MissingFields) != 0 {
		t.Fatalf("missingFields = %v, want empty", result.MissingFields)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("issues = %+v, want empty", result.Issues)
	}
}

func TestBuildOrganizationLegalDocumentVerificationManualReviewForFailedAnalysis(t *testing.T) {
	document := mustVerificationDocument(t, "ogrn_extract")
	lastError := "provider timeout"
	analysis := mustVerificationAnalysis(t, document.OrganizationID(), document.ID(), domain.OrganizationLegalDocumentAnalysisStatusFailed, json.RawMessage(`{}`), nil, nil, &lastError)

	result := domain.BuildOrganizationLegalDocumentVerification(domain.OrganizationLegalDocumentVerificationInput{
		Document: domain.OrganizationLegalDocumentVerificationDocumentInput{
			ID:             document.ID(),
			OrganizationID: document.OrganizationID().UUID(),
			DocumentType:   document.DocumentType(),
			DocumentStatus: string(document.Status()),
		},
		Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(analysis),
		Now:      time.Now().UTC(),
	})

	if result == nil {
		t.Fatal("result = nil")
	}
	if result.Verdict != domain.OrganizationLegalDocumentVerificationVerdictManualReview {
		t.Fatalf("verdict = %q, want manual_review", result.Verdict)
	}
	if len(result.Issues) != 1 || result.Issues[0].Code != "analysis_failed" {
		t.Fatalf("issues = %+v, want analysis_failed", result.Issues)
	}
}

func TestBuildOrganizationLegalDocumentVerificationRejectedOnTypeMismatch(t *testing.T) {
	document := mustVerificationDocument(t, "inn_certificate")
	analysis := mustVerificationAnalysis(t, document.OrganizationID(), document.ID(), domain.OrganizationLegalDocumentAnalysisStatusCompleted, json.RawMessage(`{"inn":"7701234567","companyName":"Acme","registrationDate":"2022-01-01"}`), stringPtr("charter"), floatPtr(0.99), nil)

	result := domain.BuildOrganizationLegalDocumentVerification(domain.OrganizationLegalDocumentVerificationInput{
		Document: domain.OrganizationLegalDocumentVerificationDocumentInput{
			ID:             document.ID(),
			OrganizationID: document.OrganizationID().UUID(),
			DocumentType:   document.DocumentType(),
			DocumentStatus: string(document.Status()),
		},
		Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(analysis),
		Now:      time.Now().UTC(),
	})

	if result == nil {
		t.Fatal("result = nil")
	}
	if result.Verdict != domain.OrganizationLegalDocumentVerificationVerdictRejected {
		t.Fatalf("verdict = %q, want rejected", result.Verdict)
	}
	if len(result.Issues) == 0 || result.Issues[0].Code != "document_type_mismatch" {
		t.Fatalf("issues = %+v, want document_type_mismatch", result.Issues)
	}
}

func mustVerificationDocument(t *testing.T, documentType string) *domain.OrganizationLegalDocument {
	t.Helper()

	orgID, err := domain.OrganizationIDFromUUID(uuid.New())
	if err != nil {
		t.Fatalf("OrganizationIDFromUUID: %v", err)
	}
	document, err := domain.RehydrateOrganizationLegalDocument(domain.RehydrateOrganizationLegalDocumentParams{
		ID:             uuid.New(),
		OrganizationID: orgID,
		DocumentType:   documentType,
		Status:         "pending",
		ObjectID:       uuid.New(),
		Title:          "Document",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      timePtr(time.Now().UTC()),
	})
	if err != nil {
		t.Fatalf("RehydrateOrganizationLegalDocument: %v", err)
	}
	return document
}

func mustVerificationAnalysis(
	t *testing.T,
	organizationID domain.OrganizationID,
	documentID uuid.UUID,
	status domain.OrganizationLegalDocumentAnalysisStatus,
	fields json.RawMessage,
	detectedType *string,
	confidence *float64,
	lastError *string,
) *domain.OrganizationLegalDocumentAnalysis {
	t.Helper()

	now := time.Now().UTC()
	analysis, err := domain.RehydrateOrganizationLegalDocumentAnalysis(domain.RehydrateOrganizationLegalDocumentAnalysisParams{
		ID:                   uuid.New(),
		DocumentID:           documentID,
		OrganizationID:       organizationID,
		Status:               string(status),
		Provider:             "generic-http",
		ExtractedFieldsJSON:  fields,
		DetectedDocumentType: detectedType,
		ConfidenceScore:      confidence,
		RequestedAt:          now,
		StartedAt:            timePtr(now),
		CompletedAt:          timePtr(now),
		UpdatedAt:            timePtr(now),
		LastError:            lastError,
	})
	if err != nil {
		t.Fatalf("RehydrateOrganizationLegalDocumentAnalysis: %v", err)
	}
	return analysis
}

func floatPtr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
