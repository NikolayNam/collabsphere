package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationLegalDocumentVerificationVerdict string

const (
	OrganizationLegalDocumentVerificationVerdictApproved     OrganizationLegalDocumentVerificationVerdict = "approved"
	OrganizationLegalDocumentVerificationVerdictManualReview OrganizationLegalDocumentVerificationVerdict = "manual_review"
	OrganizationLegalDocumentVerificationVerdictRejected     OrganizationLegalDocumentVerificationVerdict = "rejected"
)

type OrganizationLegalDocumentVerificationIssueSeverity string

const (
	OrganizationLegalDocumentVerificationIssueSeverityError   OrganizationLegalDocumentVerificationIssueSeverity = "error"
	OrganizationLegalDocumentVerificationIssueSeverityWarning OrganizationLegalDocumentVerificationIssueSeverity = "warning"
)

type OrganizationLegalDocumentVerificationIssue struct {
	Code     string
	Severity OrganizationLegalDocumentVerificationIssueSeverity
	Message  string
	Field    *string
}

type OrganizationLegalDocumentVerification struct {
	DocumentID           uuid.UUID
	OrganizationID       uuid.UUID
	DocumentType         string
	DocumentStatus       string
	AnalysisStatus       *string
	Verdict              OrganizationLegalDocumentVerificationVerdict
	Summary              string
	DetectedDocumentType *string
	ConfidenceScore      *float64
	RequiredFields       []string
	MissingFields        []string
	Issues               []OrganizationLegalDocumentVerificationIssue
	CheckedAt            time.Time
}
