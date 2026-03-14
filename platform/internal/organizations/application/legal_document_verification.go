package application

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
)

func (s *Service) GetOrganizationLegalDocumentVerification(ctx context.Context, q GetOrganizationLegalDocumentVerificationQuery) (*domain.OrganizationLegalDocumentVerification, error) {
	if err := s.requireOrganizationAccess(ctx, q.OrganizationID, q.ActorAccountID, true); err != nil {
		return nil, err
	}
	document, err := s.repo.GetOrganizationLegalDocumentByID(ctx, q.OrganizationID, q.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, fault.NotFound("Organization legal document not found")
	}
	analysis, err := s.repo.GetOrganizationLegalDocumentAnalysis(ctx, q.OrganizationID, q.DocumentID)
	if err != nil {
		return nil, err
	}
	return domain.BuildOrganizationLegalDocumentVerification(domain.OrganizationLegalDocumentVerificationInput{
		Document: domain.OrganizationLegalDocumentVerificationDocumentInput{
			ID:             document.ID(),
			OrganizationID: document.OrganizationID().UUID(),
			DocumentType:   document.DocumentType(),
			DocumentStatus: string(document.Status()),
		},
		Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(analysis),
		Now:      s.clock.Now(),
	}), nil
}

func toOrganizationLegalDocumentVerificationAnalysisInput(analysis *domain.OrganizationLegalDocumentAnalysis) *domain.OrganizationLegalDocumentVerificationAnalysisInput {
	if analysis == nil {
		return nil
	}
	return &domain.OrganizationLegalDocumentVerificationAnalysisInput{
		Status:               string(analysis.Status()),
		ExtractedFieldsJSON:  analysis.ExtractedFieldsJSON(),
		DetectedDocumentType: analysis.DetectedDocumentType(),
		ConfidenceScore:      analysis.ConfidenceScore(),
		RequestedAt:          analysis.RequestedAt(),
		CompletedAt:          analysis.CompletedAt(),
		UpdatedAt:            analysis.UpdatedAt(),
		LastError:            analysis.LastError(),
	}
}
