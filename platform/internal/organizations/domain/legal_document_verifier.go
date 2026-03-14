package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const OrganizationLegalDocumentApprovalConfidenceThreshold = 0.85

var organizationLegalDocumentVerificationRules = map[string][]string{
	"inn_certificate": {"inn", "companyName", "registrationDate"},
	"ogrn_extract":    {"ogrn", "companyName"},
	"charter":         {"companyName"},
}

type OrganizationLegalDocumentVerificationInput struct {
	Document OrganizationLegalDocumentVerificationDocumentInput
	Analysis *OrganizationLegalDocumentVerificationAnalysisInput
	Now      time.Time
}

type OrganizationLegalDocumentVerificationDocumentInput struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	DocumentType   string
	DocumentStatus string
}

type OrganizationLegalDocumentVerificationAnalysisInput struct {
	Status               string
	ExtractedFieldsJSON  json.RawMessage
	DetectedDocumentType *string
	ConfidenceScore      *float64
	RequestedAt          time.Time
	CompletedAt          *time.Time
	UpdatedAt            *time.Time
	LastError            *string
}

func BuildOrganizationLegalDocumentVerification(input OrganizationLegalDocumentVerificationInput) *OrganizationLegalDocumentVerification {
	if input.Document.ID == uuid.Nil {
		return nil
	}

	requiredFields := append([]string{}, organizationLegalDocumentVerificationRules[strings.TrimSpace(input.Document.DocumentType)]...)
	result := &OrganizationLegalDocumentVerification{
		DocumentID:     input.Document.ID,
		OrganizationID: input.Document.OrganizationID,
		DocumentType:   strings.TrimSpace(input.Document.DocumentType),
		DocumentStatus: strings.TrimSpace(input.Document.DocumentStatus),
		RequiredFields: requiredFields,
		Verdict:        OrganizationLegalDocumentVerificationVerdictManualReview,
		CheckedAt:      resolveOrganizationLegalDocumentVerificationCheckedAt(input.Analysis, input.Now),
	}

	if input.Analysis == nil {
		result.Summary = "Machine verification requires manual review because analysis is missing."
		result.Issues = []OrganizationLegalDocumentVerificationIssue{{
			Code:     "analysis_missing",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityWarning,
			Message:  "Machine analysis result is not available for this legal document.",
		}}
		return result
	}

	analysisStatus := strings.TrimSpace(input.Analysis.Status)
	result.AnalysisStatus = verificationStringPtr(analysisStatus)
	result.DetectedDocumentType = cloneStringPtr(input.Analysis.DetectedDocumentType)
	result.ConfidenceScore = cloneFloat64Ptr(input.Analysis.ConfidenceScore)

	switch OrganizationLegalDocumentAnalysisStatus(analysisStatus) {
	case OrganizationLegalDocumentAnalysisStatusPending, OrganizationLegalDocumentAnalysisStatusProcessing:
		result.Summary = "Machine verification requires manual review because analysis is still in progress."
		result.Issues = []OrganizationLegalDocumentVerificationIssue{{
			Code:     "analysis_not_ready",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityWarning,
			Message:  "Machine analysis is not completed yet.",
		}}
		return result
	case OrganizationLegalDocumentAnalysisStatusFailed:
		message := "Machine analysis failed."
		if lastError := strings.TrimSpace(derefVerificationString(input.Analysis.LastError)); lastError != "" {
			message = fmt.Sprintf("Machine analysis failed: %s", lastError)
		}
		result.Summary = "Machine verification requires manual review because analysis failed."
		result.Issues = []OrganizationLegalDocumentVerificationIssue{{
			Code:     "analysis_failed",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityError,
			Message:  message,
		}}
		return result
	}

	issues := make([]OrganizationLegalDocumentVerificationIssue, 0, 4)
	missingFields, extractedFieldsValid := findMissingOrganizationLegalDocumentVerificationFields(input.Analysis.ExtractedFieldsJSON, requiredFields)
	result.MissingFields = missingFields
	if !extractedFieldsValid {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "extracted_fields_invalid",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityError,
			Message:  "Extracted fields payload is not a JSON object.",
		})
	}
	if len(missingFields) > 0 {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "required_fields_missing",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityError,
			Message:  "Machine analysis did not extract all required document fields.",
		})
	}

	rejected := false
	manualReview := !extractedFieldsValid || len(missingFields) > 0

	detectedType := strings.TrimSpace(derefVerificationString(input.Analysis.DetectedDocumentType))
	if detectedType == "" {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "document_type_not_detected",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityWarning,
			Message:  "Machine analysis did not detect the document type.",
			Field:    verificationStringPtr("detectedDocumentType"),
		})
		manualReview = true
	} else if detectedType != result.DocumentType {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "document_type_mismatch",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityError,
			Message:  fmt.Sprintf("Detected document type %q does not match expected type %q.", detectedType, result.DocumentType),
			Field:    verificationStringPtr("detectedDocumentType"),
		})
		rejected = true
	}

	confidence := input.Analysis.ConfidenceScore
	if confidence == nil {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "confidence_missing",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityWarning,
			Message:  "Machine analysis did not return a confidence score.",
			Field:    verificationStringPtr("confidenceScore"),
		})
		manualReview = true
	} else if *confidence < OrganizationLegalDocumentApprovalConfidenceThreshold {
		issues = append(issues, OrganizationLegalDocumentVerificationIssue{
			Code:     "confidence_below_threshold",
			Severity: OrganizationLegalDocumentVerificationIssueSeverityWarning,
			Message:  fmt.Sprintf("Confidence %.2f is below the approval threshold %.2f.", *confidence, OrganizationLegalDocumentApprovalConfidenceThreshold),
			Field:    verificationStringPtr("confidenceScore"),
		})
		manualReview = true
	}

	result.Issues = issues
	switch {
	case rejected:
		result.Verdict = OrganizationLegalDocumentVerificationVerdictRejected
		result.Summary = "Machine verification rejected the document because the detected document type does not match the expected type."
	case manualReview:
		result.Verdict = OrganizationLegalDocumentVerificationVerdictManualReview
		result.Summary = "Machine verification requires manual review."
	default:
		result.Verdict = OrganizationLegalDocumentVerificationVerdictApproved
		result.Summary = "Machine verification passed."
	}
	return result
}

func findMissingOrganizationLegalDocumentVerificationFields(raw json.RawMessage, required []string) ([]string, bool) {
	if len(required) == 0 {
		return nil, true
	}
	var fields map[string]any
	if len(raw) == 0 {
		return append([]string{}, required...), false
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return append([]string{}, required...), false
	}
	if fields == nil {
		return append([]string{}, required...), false
	}

	missing := make([]string, 0, len(required))
	for _, field := range required {
		value, ok := fields[field]
		if !ok || isEmptyOrganizationLegalDocumentVerificationValue(value) {
			missing = append(missing, field)
		}
	}
	return missing, true
}

func isEmptyOrganizationLegalDocumentVerificationValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}

func resolveOrganizationLegalDocumentVerificationCheckedAt(analysis *OrganizationLegalDocumentVerificationAnalysisInput, now time.Time) time.Time {
	if analysis == nil {
		return now
	}
	if analysis.UpdatedAt != nil && !analysis.UpdatedAt.IsZero() {
		return *analysis.UpdatedAt
	}
	if analysis.CompletedAt != nil && !analysis.CompletedAt.IsZero() {
		return *analysis.CompletedAt
	}
	if !analysis.RequestedAt.IsZero() {
		return analysis.RequestedAt
	}
	return now
}

func derefVerificationString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func verificationStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
