package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func ToCooperationApplicationResponse(application *domain.CooperationApplication, status int) *dto.CooperationApplicationResponse {
	if application == nil {
		return nil
	}
	return &dto.CooperationApplicationResponse{
		Status: status,
		Body: dto.CooperationApplicationBody{
			ID:                    application.ID(),
			OrganizationID:        application.OrganizationID().UUID(),
			Status:                string(application.Status()),
			ConfirmationEmail:     application.ConfirmationEmail(),
			CompanyName:           application.CompanyName(),
			RepresentedCategories: application.RepresentedCategories(),
			MinimumOrderAmount:    application.MinimumOrderAmount(),
			DeliveryGeography:     application.DeliveryGeography(),
			SalesChannels:         application.SalesChannels(),
			StorefrontURL:         application.StorefrontURL(),
			ContactFirstName:      application.ContactFirstName(),
			ContactLastName:       application.ContactLastName(),
			ContactJobTitle:       application.ContactJobTitle(),
			PriceListObjectID:     application.PriceListObjectID(),
			ContactEmail:          application.ContactEmail(),
			ContactPhone:          application.ContactPhone(),
			PartnerCode:           application.PartnerCode(),
			ReviewNote:            application.ReviewNote(),
			ReviewerAccountID:     application.ReviewerAccountID(),
			SubmittedAt:           application.SubmittedAt(),
			ReviewedAt:            application.ReviewedAt(),
			CreatedAt:             application.CreatedAt(),
			UpdatedAt:             application.UpdatedAt(),
		},
	}
}

func ToOrganizationLegalDocumentResponse(document *domain.OrganizationLegalDocument, status int) *dto.OrganizationLegalDocumentResponse {
	if document == nil {
		return nil
	}
	return &dto.OrganizationLegalDocumentResponse{
		Status: status,
		Body:   toOrganizationLegalDocumentBody(document),
	}
}

func ToOrganizationLegalDocumentsResponse(documents []domain.OrganizationLegalDocument, status int) *dto.OrganizationLegalDocumentsResponse {
	response := &dto.OrganizationLegalDocumentsResponse{Status: status}
	response.Body.Data = make([]dto.OrganizationLegalDocumentBody, 0, len(documents))
	for i := range documents {
		response.Body.Data = append(response.Body.Data, toOrganizationLegalDocumentBody(&documents[i]))
	}
	return response
}

func ToOrganizationLegalDocumentAnalysisResponse(analysis *domain.OrganizationLegalDocumentAnalysis, status int) *dto.OrganizationLegalDocumentAnalysisResponse {
	if analysis == nil {
		return nil
	}
	return &dto.OrganizationLegalDocumentAnalysisResponse{
		Status: status,
		Body: dto.OrganizationLegalDocumentAnalysisBody{
			ID:                   analysis.ID(),
			DocumentID:           analysis.DocumentID(),
			OrganizationID:       analysis.OrganizationID().UUID(),
			Status:               string(analysis.Status()),
			Provider:             analysis.Provider(),
			ExtractedText:        analysis.ExtractedText(),
			Summary:              analysis.Summary(),
			ExtractedFields:      analysis.ExtractedFieldsJSON(),
			DetectedDocumentType: analysis.DetectedDocumentType(),
			ConfidenceScore:      analysis.ConfidenceScore(),
			RequestedAt:          analysis.RequestedAt(),
			StartedAt:            analysis.StartedAt(),
			CompletedAt:          analysis.CompletedAt(),
			UpdatedAt:            analysis.UpdatedAt(),
			LastError:            analysis.LastError(),
		},
	}
}

func toOrganizationLegalDocumentBody(document *domain.OrganizationLegalDocument) dto.OrganizationLegalDocumentBody {
	return dto.OrganizationLegalDocumentBody{
		ID:                  document.ID(),
		OrganizationID:      document.OrganizationID().UUID(),
		DocumentType:        document.DocumentType(),
		Status:              string(document.Status()),
		ObjectID:            document.ObjectID(),
		Title:               document.Title(),
		UploadedByAccountID: document.UploadedByAccountID(),
		ReviewerAccountID:   document.ReviewerAccountID(),
		ReviewNote:          document.ReviewNote(),
		CreatedAt:           document.CreatedAt(),
		UpdatedAt:           document.UpdatedAt(),
		ReviewedAt:          document.ReviewedAt(),
	}
}
