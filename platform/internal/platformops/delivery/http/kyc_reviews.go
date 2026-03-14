package http

import (
	"context"
	"net/http"

	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	platformdto "github.com/NikolayNam/collabsphere/internal/platformops/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	"github.com/google/uuid"
)

func (h *Handler) ListKYCReviews(ctx context.Context, input *platformdto.ListKYCReviewsInput) (*platformdto.KYCReviewListResponse, error) {
	items, total, err := h.svc.ListKYCReviews(ctx, platformapp.ListKYCReviewsCmd{
		Scope:  optionalString(input.Scope),
		Status: optionalString(input.Status),
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.KYCReviewListResponse{Status: http.StatusOK}
	out.Body.Total = total
	out.Body.Items = make([]platformdto.KYCReviewItem, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, toKYCReviewItemDTO(item))
	}
	return out, nil
}

func (h *Handler) GetKYCReview(ctx context.Context, input *platformdto.GetKYCReviewInput) (*platformdto.KYCReviewResponse, error) {
	detail, err := h.svc.GetKYCReview(ctx, platformapp.GetKYCReviewCmd{
		ReviewID: input.ReviewID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.KYCReviewResponse{Status: http.StatusOK}
	out.Body = toKYCReviewDetailDTO(*detail)
	return out, nil
}

func (h *Handler) DecideKYCReview(ctx context.Context, input *platformdto.DecideKYCReviewInput) (*platformdto.KYCReviewResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	detail, err := h.svc.DecideKYCReview(ctx, platformapp.DecideKYCReviewCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		ReviewID:       input.ReviewID,
		Decision:       input.Body.Decision,
		Reason:         input.Body.Reason,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.KYCReviewResponse{Status: http.StatusOK}
	out.Body = toKYCReviewDetailDTO(*detail)
	return out, nil
}

func (h *Handler) DecideKYCDocument(ctx context.Context, input *platformdto.DecideKYCDocumentInput) (*platformdto.KYCReviewResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	documentID, err := uuid.Parse(input.DocumentID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Document id is invalid", fault.Field("documentId", "must be UUID")))
	}
	detail, err := h.svc.DecideKYCDocumentReview(ctx, platformapp.DecideKYCDocumentReviewCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		ReviewID:       input.ReviewID,
		DocumentID:     documentID,
		Decision:       input.Body.Decision,
		Reason:         input.Body.Reason,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.KYCReviewResponse{Status: http.StatusOK}
	out.Body = toKYCReviewDetailDTO(*detail)
	return out, nil
}

func (h *Handler) ListKYCLevels(ctx context.Context, input *platformdto.ListKYCLevelsInput) (*platformdto.KYCLevelListResponse, error) {
	items, err := h.svc.ListKYCLevels(ctx, platformapp.ListKYCLevelsCmd{
		Scope: optionalString(input.Scope),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.KYCLevelListResponse{Status: http.StatusOK}
	out.Body.Items = make([]platformdto.KYCLevel, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, toKYCLevelDTO(item))
	}
	return out, nil
}

func (h *Handler) CreateKYCLevel(ctx context.Context, input *platformdto.CreateKYCLevelInput) (*platformdto.KYCLevelResponse, error) {
	item, err := h.svc.UpsertKYCLevel(ctx, platformapp.UpsertKYCLevelCmd{
		Scope:                 input.Body.Scope,
		Code:                  input.Body.Code,
		Name:                  input.Body.Name,
		Rank:                  input.Body.Rank,
		IsActive:              input.Body.IsActive,
		RequiredDocumentTypes: toDomainLevelRequirements(input.Body.RequiredDocumentTypes),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &platformdto.KYCLevelResponse{
		Status: http.StatusOK,
		Body:   toKYCLevelDTO(*item),
	}, nil
}

func (h *Handler) UpdateKYCLevel(ctx context.Context, input *platformdto.UpdateKYCLevelInput) (*platformdto.KYCLevelResponse, error) {
	levelID, err := uuid.Parse(input.LevelID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("KYC level id is invalid", fault.Field("levelId", "must be UUID")))
	}
	item, err := h.svc.UpsertKYCLevel(ctx, platformapp.UpsertKYCLevelCmd{
		ID:                    &levelID,
		Scope:                 input.Body.Scope,
		Code:                  input.Body.Code,
		Name:                  input.Body.Name,
		Rank:                  input.Body.Rank,
		IsActive:              input.Body.IsActive,
		RequiredDocumentTypes: toDomainLevelRequirements(input.Body.RequiredDocumentTypes),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &platformdto.KYCLevelResponse{
		Status: http.StatusOK,
		Body:   toKYCLevelDTO(*item),
	}, nil
}

func (h *Handler) DeleteKYCLevel(ctx context.Context, input *platformdto.DeleteKYCLevelInput) (*struct{ Status int }, error) {
	levelID, err := uuid.Parse(input.LevelID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("KYC level id is invalid", fault.Field("levelId", "must be UUID")))
	}
	if err := h.svc.DeleteKYCLevel(ctx, platformapp.DeleteKYCLevelCmd{LevelID: levelID}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &struct{ Status int }{Status: http.StatusNoContent}, nil
}

func (h *Handler) IssueKYCLevel(ctx context.Context, input *platformdto.IssueKYCLevelInput) (*platformdto.IssueKYCLevelResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	assignment, err := h.svc.IssueKYCLevel(ctx, platformapp.IssueKYCLevelCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		ReviewID:       input.ReviewID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.IssueKYCLevelResponse{Status: http.StatusOK}
	if assignment != nil {
		out.Body.LevelID = assignment.LevelID
		out.Body.LevelCode = assignment.LevelCode
		out.Body.LevelName = assignment.LevelName
		out.Body.IssuedAt = assignment.IssuedAt
	}
	return out, nil
}

func toKYCReviewItemDTO(item domain.KYCReviewItem) platformdto.KYCReviewItem {
	return platformdto.KYCReviewItem{
		ReviewID:     item.ReviewID,
		Scope:        item.Scope,
		SubjectID:    item.SubjectID,
		Status:       item.Status,
		KYCLevelCode: item.KYCLevelCode,
		KYCLevelName: item.KYCLevelName,
		LegalName:    item.LegalName,
		CountryCode:  item.CountryCode,
		SubmittedAt:  item.SubmittedAt,
		ReviewedAt:   item.ReviewedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func toKYCReviewDetailDTO(item domain.KYCReviewDetail) platformdto.KYCReviewItem {
	createdAt := item.CreatedAt
	documents := make([]platformdto.KYCDocumentReviewItem, 0, len(item.Documents))
	for _, document := range item.Documents {
		documents = append(documents, platformdto.KYCDocumentReviewItem{
			ID:                document.ID,
			ObjectID:          document.ObjectID,
			DocumentType:      document.DocumentType,
			Title:             document.Title,
			Status:            document.Status,
			ReviewNote:        document.ReviewNote,
			ReviewerAccountID: document.ReviewerAccountID,
			CreatedAt:         document.CreatedAt,
			UpdatedAt:         document.UpdatedAt,
			ReviewedAt:        document.ReviewedAt,
		})
	}
	events := make([]platformdto.KYCReviewEvent, 0, len(item.Events))
	for _, event := range item.Events {
		events = append(events, platformdto.KYCReviewEvent{
			ID:                event.ID,
			Scope:             event.Scope,
			SubjectID:         event.SubjectID,
			Decision:          event.Decision,
			Reason:            event.Reason,
			ReviewerAccountID: event.ReviewerAccountID,
			CreatedAt:         event.CreatedAt,
		})
	}
	return platformdto.KYCReviewItem{
		ReviewID:           item.ReviewID,
		Scope:              item.Scope,
		SubjectID:          item.SubjectID,
		Status:             item.Status,
		KYCLevelCode:       item.KYCLevelCode,
		KYCLevelName:       item.KYCLevelName,
		LegalName:          item.LegalName,
		CountryCode:        item.CountryCode,
		RegistrationNumber: item.RegistrationNumber,
		TaxID:              item.TaxID,
		DocumentNumber:     item.DocumentNumber,
		ResidenceAddress:   item.ResidenceAddress,
		ReviewNote:         item.ReviewNote,
		ReviewerAccountID:  item.ReviewerAccountID,
		SubmittedAt:        item.SubmittedAt,
		ReviewedAt:         item.ReviewedAt,
		CreatedAt:          &createdAt,
		UpdatedAt:          item.UpdatedAt,
		Documents:          documents,
		Events:             events,
	}
}

func toKYCLevelDTO(item domain.KYCLevel) platformdto.KYCLevel {
	requirements := make([]platformdto.KYCLevelRequirement, 0, len(item.RequiredDocumentTypes))
	for _, requirement := range item.RequiredDocumentTypes {
		requirements = append(requirements, platformdto.KYCLevelRequirement{
			DocumentType: requirement.DocumentType,
			MinCount:     requirement.MinCount,
		})
	}
	return platformdto.KYCLevel{
		ID:                    item.ID,
		Scope:                 item.Scope,
		Code:                  item.Code,
		Name:                  item.Name,
		Rank:                  item.Rank,
		IsActive:              item.IsActive,
		RequiredDocumentTypes: requirements,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}

func toDomainLevelRequirements(items []platformdto.KYCLevelRequirement) []domain.KYCLevelRequirement {
	requirements := make([]domain.KYCLevelRequirement, 0, len(items))
	for _, item := range items {
		requirements = append(requirements, domain.KYCLevelRequirement{
			DocumentType: item.DocumentType,
			MinCount:     item.MinCount,
		})
	}
	return requirements
}
