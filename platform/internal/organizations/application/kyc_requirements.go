package application

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func (s *Service) GetOrganizationKYCRequirements(ctx context.Context, q GetOrganizationKYCRequirementsQuery) (*domain.OrganizationKYCRequirements, error) {
	if err := s.requireOrganizationAccess(ctx, q.OrganizationID, q.ActorAccountID, true); err != nil {
		return nil, err
	}

	application, err := s.repo.GetCooperationApplication(ctx, q.OrganizationID)
	if err != nil {
		return nil, err
	}
	documents, err := s.repo.ListOrganizationLegalDocuments(ctx, q.OrganizationID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	legalDocuments, err := s.loadOrganizationKYCLegalDocuments(ctx, q.OrganizationID, documents, now)
	if err != nil {
		return nil, err
	}

	return domain.BuildOrganizationKYCRequirements(domain.OrganizationKYCRequirementsInput{
		OrganizationID:         q.OrganizationID.UUID(),
		CooperationApplication: toOrganizationKYCCooperationApplicationInput(application),
		LegalDocuments:         legalDocuments,
		Now:                    now,
	}), nil
}

func (s *Service) loadOrganizationKYCLegalDocuments(ctx context.Context, organizationID domain.OrganizationID, documents []domain.OrganizationLegalDocument, now time.Time) ([]domain.OrganizationKYCLegalDocumentInput, error) {
	if len(documents) == 0 {
		return nil, nil
	}
	items := make([]domain.OrganizationKYCLegalDocumentInput, 0, len(documents))
	for i := range documents {
		document := documents[i]
		analysis, err := s.repo.GetOrganizationLegalDocumentAnalysis(ctx, organizationID, document.ID())
		if err != nil {
			return nil, err
		}
		items = append(items, domain.OrganizationKYCLegalDocumentInput{
			ID:           document.ID(),
			DocumentType: document.DocumentType(),
			Status:       string(document.Status()),
			ReviewNote:   document.ReviewNote(),
			Verification: domain.BuildOrganizationLegalDocumentVerification(domain.OrganizationLegalDocumentVerificationInput{
				Document: domain.OrganizationLegalDocumentVerificationDocumentInput{
					ID:             document.ID(),
					OrganizationID: document.OrganizationID().UUID(),
					DocumentType:   document.DocumentType(),
					DocumentStatus: string(document.Status()),
				},
				Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(analysis),
				Now:      now,
			}),
		})
	}
	return items, nil
}

func toOrganizationKYCCooperationApplicationInput(application *domain.CooperationApplication) *domain.OrganizationKYCCooperationApplicationInput {
	if application == nil {
		return nil
	}
	return &domain.OrganizationKYCCooperationApplicationInput{
		Status:                string(application.Status()),
		ReviewNote:            application.ReviewNote(),
		ConfirmationEmail:     application.ConfirmationEmail(),
		CompanyName:           application.CompanyName(),
		RepresentedCategories: application.RepresentedCategories(),
		MinimumOrderAmount:    application.MinimumOrderAmount(),
		DeliveryGeography:     application.DeliveryGeography(),
		SalesChannels:         application.SalesChannels(),
		PriceListObjectID:     application.PriceListObjectID(),
		ContactFirstName:      application.ContactFirstName(),
		ContactLastName:       application.ContactLastName(),
		ContactJobTitle:       application.ContactJobTitle(),
		ContactEmail:          application.ContactEmail(),
		ContactPhone:          application.ContactPhone(),
	}
}
