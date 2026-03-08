package domain

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type OrganizationLegalDocumentAnalysisStatus string

const (
	OrganizationLegalDocumentAnalysisStatusPending    OrganizationLegalDocumentAnalysisStatus = "pending"
	OrganizationLegalDocumentAnalysisStatusProcessing OrganizationLegalDocumentAnalysisStatus = "processing"
	OrganizationLegalDocumentAnalysisStatusCompleted  OrganizationLegalDocumentAnalysisStatus = "completed"
	OrganizationLegalDocumentAnalysisStatusFailed     OrganizationLegalDocumentAnalysisStatus = "failed"
)

type OrganizationLegalDocumentAnalysis struct {
	id                   uuid.UUID
	documentID           uuid.UUID
	organizationID       OrganizationID
	status               OrganizationLegalDocumentAnalysisStatus
	provider             string
	extractedText        *string
	summary              *string
	extractedFieldsJSON  json.RawMessage
	detectedDocumentType *string
	confidenceScore      *float64
	requestedAt          time.Time
	startedAt            *time.Time
	completedAt          *time.Time
	updatedAt            *time.Time
	lastError            *string
}

type RehydrateOrganizationLegalDocumentAnalysisParams struct {
	ID                   uuid.UUID
	DocumentID           uuid.UUID
	OrganizationID       OrganizationID
	Status               string
	Provider             string
	ExtractedText        *string
	Summary              *string
	ExtractedFieldsJSON  json.RawMessage
	DetectedDocumentType *string
	ConfidenceScore      *float64
	RequestedAt          time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
	UpdatedAt            *time.Time
	LastError            *string
}

func RehydrateOrganizationLegalDocumentAnalysis(p RehydrateOrganizationLegalDocumentAnalysisParams) (*OrganizationLegalDocumentAnalysis, error) {
	if p.ID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentAnalysisIDEmpty
	}
	if p.DocumentID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.RequestedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	status, err := normalizeOrganizationLegalDocumentAnalysisStatus(p.Status)
	if err != nil {
		return nil, err
	}
	provider, err := normalizeOrganizationLegalDocumentAnalysisProvider(p.Provider)
	if err != nil {
		return nil, err
	}
	extractedText, err := normalizeOptionalOrgField(p.ExtractedText, 1000000, ErrOrganizationLegalDocumentAnalysisExtractedTextInvalid)
	if err != nil {
		return nil, err
	}
	summary, err := normalizeOptionalOrgField(p.Summary, 4096, ErrOrganizationLegalDocumentAnalysisSummaryInvalid)
	if err != nil {
		return nil, err
	}
	detectedDocumentType, err := normalizeOptionalOrgField(p.DetectedDocumentType, 128, ErrOrganizationLegalDocumentAnalysisDetectedTypeInvalid)
	if err != nil {
		return nil, err
	}
	lastError, err := normalizeOptionalOrgField(p.LastError, 4096, ErrOrganizationLegalDocumentAnalysisLastErrorInvalid)
	if err != nil {
		return nil, err
	}
	fieldsJSON := p.ExtractedFieldsJSON
	if len(fieldsJSON) == 0 {
		fieldsJSON = json.RawMessage(`{}`)
	}
	if !json.Valid(fieldsJSON) {
		return nil, ErrOrganizationLegalDocumentAnalysisFieldsInvalid
	}
	if p.ConfidenceScore != nil && (*p.ConfidenceScore < 0 || *p.ConfidenceScore > 1) {
		return nil, ErrOrganizationLegalDocumentAnalysisConfidenceInvalid
	}
	return &OrganizationLegalDocumentAnalysis{
		id:                   p.ID,
		documentID:           p.DocumentID,
		organizationID:       p.OrganizationID,
		status:               status,
		provider:             provider,
		extractedText:        extractedText,
		summary:              summary,
		extractedFieldsJSON:  fieldsJSON,
		detectedDocumentType: detectedDocumentType,
		confidenceScore:      cloneFloat64Ptr(p.ConfidenceScore),
		requestedAt:          p.RequestedAt,
		startedAt:            cloneTimePtr(p.StartedAt),
		completedAt:          cloneTimePtr(p.CompletedAt),
		updatedAt:            cloneTimePtr(p.UpdatedAt),
		lastError:            lastError,
	}, nil
}

func (a *OrganizationLegalDocumentAnalysis) ID() uuid.UUID                  { return a.id }
func (a *OrganizationLegalDocumentAnalysis) DocumentID() uuid.UUID          { return a.documentID }
func (a *OrganizationLegalDocumentAnalysis) OrganizationID() OrganizationID { return a.organizationID }
func (a *OrganizationLegalDocumentAnalysis) Status() OrganizationLegalDocumentAnalysisStatus {
	return a.status
}
func (a *OrganizationLegalDocumentAnalysis) Provider() string { return a.provider }
func (a *OrganizationLegalDocumentAnalysis) ExtractedText() *string {
	return cloneStringPtr(a.extractedText)
}
func (a *OrganizationLegalDocumentAnalysis) Summary() *string { return cloneStringPtr(a.summary) }
func (a *OrganizationLegalDocumentAnalysis) ExtractedFieldsJSON() json.RawMessage {
	return cloneJSONRawMessage(a.extractedFieldsJSON)
}
func (a *OrganizationLegalDocumentAnalysis) DetectedDocumentType() *string {
	return cloneStringPtr(a.detectedDocumentType)
}
func (a *OrganizationLegalDocumentAnalysis) ConfidenceScore() *float64 {
	return cloneFloat64Ptr(a.confidenceScore)
}
func (a *OrganizationLegalDocumentAnalysis) RequestedAt() time.Time { return a.requestedAt }
func (a *OrganizationLegalDocumentAnalysis) StartedAt() *time.Time  { return cloneTimePtr(a.startedAt) }
func (a *OrganizationLegalDocumentAnalysis) CompletedAt() *time.Time {
	return cloneTimePtr(a.completedAt)
}
func (a *OrganizationLegalDocumentAnalysis) UpdatedAt() *time.Time { return cloneTimePtr(a.updatedAt) }
func (a *OrganizationLegalDocumentAnalysis) LastError() *string    { return cloneStringPtr(a.lastError) }

func normalizeOrganizationLegalDocumentAnalysisStatus(value string) (OrganizationLegalDocumentAnalysisStatus, error) {
	switch OrganizationLegalDocumentAnalysisStatus(strings.TrimSpace(value)) {
	case OrganizationLegalDocumentAnalysisStatusPending,
		OrganizationLegalDocumentAnalysisStatusProcessing,
		OrganizationLegalDocumentAnalysisStatusCompleted,
		OrganizationLegalDocumentAnalysisStatusFailed:
		return OrganizationLegalDocumentAnalysisStatus(strings.TrimSpace(value)), nil
	default:
		return "", ErrOrganizationLegalDocumentAnalysisStatusInvalid
	}
}

func normalizeOrganizationLegalDocumentAnalysisProvider(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
		return "", ErrOrganizationLegalDocumentAnalysisProviderInvalid
	}
	return trimmed, nil
}

func cloneFloat64Ptr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func cloneJSONRawMessage(v json.RawMessage) json.RawMessage {
	if len(v) == 0 {
		return json.RawMessage(`{}`)
	}
	out := make(json.RawMessage, len(v))
	copy(out, v)
	return out
}
