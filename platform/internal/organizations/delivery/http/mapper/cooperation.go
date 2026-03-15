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
			PriceListStatus:       string(application.PriceListStatus()),
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

func ToOrganizationLegalDocumentVerificationResponse(verification *domain.OrganizationLegalDocumentVerification, status int) *dto.OrganizationLegalDocumentVerificationResponse {
	if verification == nil {
		return nil
	}
	response := &dto.OrganizationLegalDocumentVerificationResponse{
		Status: status,
		Body: dto.OrganizationLegalDocumentVerificationBody{
			DocumentID:           verification.DocumentID,
			OrganizationID:       verification.OrganizationID,
			DocumentType:         verification.DocumentType,
			DocumentStatus:       verification.DocumentStatus,
			AnalysisStatus:       verification.AnalysisStatus,
			Verdict:              string(verification.Verdict),
			Summary:              verification.Summary,
			DetectedDocumentType: verification.DetectedDocumentType,
			ConfidenceScore:      verification.ConfidenceScore,
			RequiredFields:       append([]string{}, verification.RequiredFields...),
			MissingFields:        append([]string{}, verification.MissingFields...),
			CheckedAt:            verification.CheckedAt,
		},
	}
	response.Body.Issues = make([]dto.OrganizationLegalDocumentVerificationIssueBody, 0, len(verification.Issues))
	for _, issue := range verification.Issues {
		response.Body.Issues = append(response.Body.Issues, dto.OrganizationLegalDocumentVerificationIssueBody{
			Code:     issue.Code,
			Severity: string(issue.Severity),
			Message:  issue.Message,
			Field:    issue.Field,
		})
	}
	return response
}

func ToOrganizationKYCRequirementsResponse(requirements *domain.OrganizationKYCRequirements, status int) *dto.OrganizationKYCRequirementsResponse {
	if requirements == nil {
		return nil
	}
	response := &dto.OrganizationKYCRequirementsResponse{
		Status: status,
		Body: dto.OrganizationKYCRequirementsBody{
			OrganizationID: requirements.OrganizationID,
			Status:         string(requirements.Status),
			DisabledReason: requirements.DisabledReason,
			CheckedAt:      requirements.CheckedAt,
		},
	}
	response.Body.CurrentlyDue = toOrganizationKYCRequirementItemBodies(requirements.CurrentlyDue)
	response.Body.PendingVerification = toOrganizationKYCRequirementItemBodies(requirements.PendingVerification)
	response.Body.EventuallyDue = toOrganizationKYCRequirementItemBodies(requirements.EventuallyDue)
	response.Body.Errors = toOrganizationKYCRequirementItemBodies(requirements.Errors)
	return response
}

func toOrganizationKYCRequirementItemBodies(items []domain.OrganizationKYCRequirementItem) []dto.OrganizationKYCRequirementItemBody {
	if len(items) == 0 {
		return []dto.OrganizationKYCRequirementItemBody{}
	}
	out := make([]dto.OrganizationKYCRequirementItemBody, 0, len(items))
	for _, item := range items {
		out = append(out, dto.OrganizationKYCRequirementItemBody{
			Code:         item.Code,
			Category:     string(item.Category),
			Title:        item.Title,
			Description:  item.Description,
			Field:        item.Field,
			DocumentID:   item.DocumentID,
			DocumentType: item.DocumentType,
			Reason:       item.Reason,
		})
	}
	return out
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
