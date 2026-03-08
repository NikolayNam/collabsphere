package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDomainOrganizationLegalDocumentAnalysis(m *dbmodel.OrganizationLegalDocumentAnalysis) (*domain.OrganizationLegalDocumentAnalysis, error) {
	if m == nil {
		return nil, nil
	}
	organizationID, err := domain.OrganizationIDFromUUID(m.OrganizationID)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateOrganizationLegalDocumentAnalysis(domain.RehydrateOrganizationLegalDocumentAnalysisParams{
		ID:                   m.ID,
		DocumentID:           m.DocumentID,
		OrganizationID:       organizationID,
		Status:               m.Status,
		Provider:             m.Provider,
		ExtractedText:        m.ExtractedText,
		Summary:              m.Summary,
		ExtractedFieldsJSON:  m.ExtractedFieldsJSON,
		DetectedDocumentType: m.DetectedDocumentType,
		ConfidenceScore:      m.ConfidenceScore,
		RequestedAt:          m.RequestedAt,
		StartedAt:            m.StartedAt,
		CompletedAt:          m.CompletedAt,
		UpdatedAt:            m.UpdatedAt,
		LastError:            m.LastError,
	})
}
