package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	platformports "github.com/NikolayNam/collabsphere/internal/platformops/application/ports"
	platformdto "github.com/NikolayNam/collabsphere/internal/platformops/delivery/http/dto"
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	"github.com/google/uuid"
)

type Handler struct {
	svc *platformapp.Service
}

func NewHandler(svc *platformapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetMyAccess(ctx context.Context, _ *platformdto.AccessMeInput) (*platformdto.PlatformAccessResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	return accessResponse(http.StatusOK, access), nil
}

func (h *Handler) GetAccountRoles(ctx context.Context, input *platformdto.GetAccountRolesInput) (*platformdto.PlatformAccessResponse, error) {
	accountID, err := httpbind.ParseUUID(input.AccountID, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	access, err := h.svc.GetAccountAccess(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return accessResponse(http.StatusOK, access), nil
}

func (h *Handler) ReplaceAccountRoles(ctx context.Context, input *platformdto.ReplaceAccountRolesInput) (*platformdto.PlatformAccessResponse, error) {
	accountID, err := httpbind.ParseUUID(input.AccountID, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	result, err := h.svc.ReplaceAccountRoles(ctx, platformapp.ReplaceAccountRolesCmd{
		ActorAccountID:  access.AccountID,
		ActorRoles:      access.EffectiveRoles,
		ActorBootstrap:  access.BootstrapAdmin,
		TargetAccountID: accountID,
		Roles:           input.Body.Roles,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return accessResponse(http.StatusOK, result), nil
}

func (h *Handler) ListAutoGrantRules(ctx context.Context, _ *platformdto.ListAutoGrantRulesInput) (*platformdto.AutoGrantRuleListResponse, error) {
	rules, err := h.svc.ListAutoGrantRules(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AutoGrantRuleListResponse{Status: http.StatusOK}
	out.Body.Items = make([]platformdto.AutoGrantRule, 0, len(rules))
	for _, rule := range rules {
		out.Body.Items = append(out.Body.Items, autoGrantRuleDTO(rule))
	}
	return out, nil
}

func (h *Handler) CreateAutoGrantRule(ctx context.Context, input *platformdto.CreateAutoGrantRuleInput) (*platformdto.AutoGrantRuleResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	rule, err := h.svc.AddAutoGrantRule(ctx, platformapp.AddAutoGrantRuleCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		Role:           input.Body.Role,
		MatchType:      input.Body.MatchType,
		MatchValue:     input.Body.MatchValue,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AutoGrantRuleResponse{Status: http.StatusCreated}
	out.Body = autoGrantRuleDTO(*rule)
	return out, nil
}

func (h *Handler) DeleteAutoGrantRule(ctx context.Context, input *platformdto.DeleteAutoGrantRuleInput) (*platformdto.AutoGrantRuleResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	ruleID, err := httpbind.ParseUUID(input.RuleID, fault.Validation("Auto-grant rule id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("ruleId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	rule, err := h.svc.DeleteAutoGrantRule(ctx, platformapp.DeleteAutoGrantRuleCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		RuleID:         ruleID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AutoGrantRuleResponse{Status: http.StatusOK}
	out.Body = autoGrantRuleDTO(*rule)
	return out, nil
}

func (h *Handler) GetDashboardSummary(ctx context.Context, _ *platformdto.DashboardSummaryInput) (*platformdto.DashboardSummaryResponse, error) {
	summary, err := h.svc.GetDashboardSummary(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.DashboardSummaryResponse{Status: http.StatusOK}
	out.Body.TotalAccounts = summary.TotalAccounts
	out.Body.ActiveAccounts = summary.ActiveAccounts
	out.Body.TotalOrganizations = summary.TotalOrganizations
	out.Body.ActiveOrganizations = summary.ActiveOrganizations
	out.Body.PendingUploads = summary.PendingUploads
	out.Body.ReadyUploads = summary.ReadyUploads
	out.Body.FailedUploads = summary.FailedUploads
	out.Body.CooperationDraft = summary.CooperationDraft
	out.Body.CooperationSubmitted = summary.CooperationSubmitted
	out.Body.CooperationUnderReview = summary.CooperationUnderReview
	out.Body.CooperationApproved = summary.CooperationApproved
	out.Body.CooperationRejected = summary.CooperationRejected
	out.Body.CooperationNeedsInfo = summary.CooperationNeedsInfo
	return out, nil
}

func (h *Handler) ListAttachmentLimits(ctx context.Context, input *platformdto.ListAttachmentLimitsInput) (*platformdto.AttachmentLimitListResponse, error) {
	var scopeType *string
	if s := strings.TrimSpace(input.ScopeType); s != "" {
		scopeType = &s
	}
	var scopeID *uuid.UUID
	if s := strings.TrimSpace(input.ScopeID); s != "" {
		parsed, err := httpbind.ParseUUID(s, fault.Validation("scopeId is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("scopeId", "must be a UUID")))
		if err != nil {
			return nil, humaerr.From(ctx, err)
		}
		scopeID = &parsed
	}
	limits, err := h.svc.ListAttachmentLimits(ctx, scopeType, scopeID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitListResponse{Status: http.StatusOK}
	out.Body.Items = make([]platformdto.AttachmentLimit, 0, len(limits))
	for _, l := range limits {
		out.Body.Items = append(out.Body.Items, attachmentLimitDTO(l))
	}
	return out, nil
}

func (h *Handler) GetPlatformAttachmentLimit(ctx context.Context, _ *platformdto.GetPlatformAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	limit, err := h.svc.GetPlatformAttachmentLimit(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) UpsertPlatformAttachmentLimit(ctx context.Context, input *platformdto.UpsertPlatformAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	limit, err := h.svc.UpsertPlatformAttachmentLimit(ctx, platformports.AttachmentLimit{
		DocumentLimitBytes: input.Body.DocumentLimitBytes,
		PhotoLimitBytes:    input.Body.PhotoLimitBytes,
		VideoLimitBytes:    input.Body.VideoLimitBytes,
		TotalLimitBytes:    input.Body.TotalLimitBytes,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) GetOrganizationAttachmentLimit(ctx context.Context, input *platformdto.GetOrganizationAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	limit, err := h.svc.GetOrganizationAttachmentLimit(ctx, organizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) UpsertOrganizationAttachmentLimit(ctx context.Context, input *platformdto.UpsertOrganizationAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	limit, err := h.svc.UpsertOrganizationAttachmentLimit(ctx, organizationID, platformports.AttachmentLimit{
		DocumentLimitBytes: input.Body.DocumentLimitBytes,
		PhotoLimitBytes:    input.Body.PhotoLimitBytes,
		VideoLimitBytes:    input.Body.VideoLimitBytes,
		TotalLimitBytes:    input.Body.TotalLimitBytes,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) DeleteOrganizationAttachmentLimit(ctx context.Context, input *platformdto.DeleteOrganizationAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.DeleteOrganizationAttachmentLimit(ctx, organizationID); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &platformdto.AttachmentLimitResponse{Status: http.StatusNoContent}, nil
}

func (h *Handler) GetAccountAttachmentLimit(ctx context.Context, input *platformdto.GetAccountAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	accountID, err := httpbind.ParseUUID(input.AccountID, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	limit, err := h.svc.GetAccountAttachmentLimit(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) UpsertAccountAttachmentLimit(ctx context.Context, input *platformdto.UpsertAccountAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	accountID, err := httpbind.ParseUUID(input.AccountID, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	limit, err := h.svc.UpsertAccountAttachmentLimit(ctx, accountID, platformports.AttachmentLimit{
		DocumentLimitBytes: input.Body.DocumentLimitBytes,
		PhotoLimitBytes:    input.Body.PhotoLimitBytes,
		VideoLimitBytes:    input.Body.VideoLimitBytes,
		TotalLimitBytes:    input.Body.TotalLimitBytes,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.AttachmentLimitResponse{Status: http.StatusOK}
	out.Body = attachmentLimitDTO(*limit)
	return out, nil
}

func (h *Handler) DeleteAccountAttachmentLimit(ctx context.Context, input *platformdto.DeleteAccountAttachmentLimitInput) (*platformdto.AttachmentLimitResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	accountID, err := httpbind.ParseUUID(input.AccountID, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.DeleteAccountAttachmentLimit(ctx, accountID); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &platformdto.AttachmentLimitResponse{Status: http.StatusNoContent}, nil
}

func attachmentLimitDTO(l platformports.AttachmentLimit) platformdto.AttachmentLimit {
	return platformdto.AttachmentLimit{
		ID:                 l.ID,
		ScopeType:          l.ScopeType,
		ScopeID:            l.ScopeID,
		DocumentLimitBytes: l.DocumentLimitBytes,
		PhotoLimitBytes:    l.PhotoLimitBytes,
		VideoLimitBytes:    l.VideoLimitBytes,
		TotalLimitBytes:    l.TotalLimitBytes,
		CreatedAt:          l.CreatedAt,
		UpdatedAt:          l.UpdatedAt,
	}
}

func (h *Handler) ListUploads(ctx context.Context, input *platformdto.ListUploadsInput) (*platformdto.UploadQueueResponse, error) {
	organizationID, err := parseOptionalUUID(input.OrganizationID, "organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	createdByAccountID, err := parseOptionalUUID(input.CreatedByAccountID, "createdByAccountId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, total, err := h.svc.ListUploadQueue(ctx, platformapp.ListUploadQueueCmd{
		Status:             optionalString(input.Status),
		Purpose:            optionalString(input.Purpose),
		OrganizationID:     organizationID,
		CreatedByAccountID: createdByAccountID,
		Limit:              input.Limit,
		Offset:             input.Offset,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.UploadQueueResponse{Status: http.StatusOK}
	out.Body.Total = total
	out.Body.Items = make([]platformdto.UploadQueueItem, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, platformdto.UploadQueueItem{
			ID:                 item.ID,
			OrganizationID:     item.OrganizationID,
			CreatedByAccountID: item.CreatedByAccountID,
			Purpose:            item.Purpose,
			Status:             item.Status,
			FileName:           item.FileName,
			ContentType:        item.ContentType,
			DeclaredSizeBytes:  item.DeclaredSizeBytes,
			ErrorCode:          item.ErrorCode,
			ErrorMessage:       item.ErrorMessage,
			ResultKind:         item.ResultKind,
			ResultID:           item.ResultID,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		})
	}
	return out, nil
}

func (h *Handler) ListOrganizationReviews(ctx context.Context, input *platformdto.ListOrganizationReviewsInput) (*platformdto.OrganizationReviewQueueResponse, error) {
	organizationID, err := parseOptionalUUID(input.OrganizationID, "organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	reviewerAccountID, err := parseOptionalUUID(input.ReviewerAccountID, "reviewerAccountId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, total, err := h.svc.ListOrganizationReviewQueue(ctx, platformapp.ListOrganizationReviewQueueCmd{
		Status:            optionalString(input.Status),
		OrganizationID:    organizationID,
		ReviewerAccountID: reviewerAccountID,
		Search:            optionalString(input.Q),
		Limit:             input.Limit,
		Offset:            input.Offset,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.OrganizationReviewQueueResponse{Status: http.StatusOK}
	out.Body.Total = total
	out.Body.Items = make([]platformdto.OrganizationReviewQueueItem, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, organizationReviewQueueItemDTO(item))
	}
	return out, nil
}

func (h *Handler) GetOrganizationReview(ctx context.Context, input *platformdto.GetOrganizationReviewInput) (*platformdto.OrganizationReviewDetailResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	detail, err := h.svc.GetOrganizationReview(ctx, organizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.OrganizationReviewDetailResponse{Status: http.StatusOK}
	out.Body.Organization = organizationReviewOrganizationDTO(detail.Organization)
	out.Body.Domains = make([]platformdto.OrganizationReviewDomain, 0, len(detail.Domains))
	for _, item := range detail.Domains {
		out.Body.Domains = append(out.Body.Domains, organizationReviewDomainDTO(item))
	}
	out.Body.CooperationApplication = organizationReviewCooperationApplicationDTO(detail.CooperationApplication)
	out.Body.LegalDocuments = make([]platformdto.OrganizationReviewLegalDocument, 0, len(detail.LegalDocuments))
	for _, item := range detail.LegalDocuments {
		out.Body.LegalDocuments = append(out.Body.LegalDocuments, organizationReviewLegalDocumentDTO(item))
	}
	out.Body.KYC = organizationReviewKYCDTO(detail.KYC)
	return out, nil
}

func (h *Handler) TransitionCooperationApplicationReview(ctx context.Context, input *platformdto.TransitionCooperationApplicationReviewInput) (*platformdto.CooperationApplicationReviewResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	result, err := h.svc.TransitionCooperationApplicationReview(ctx, platformapp.TransitionCooperationApplicationReviewCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		OrganizationID: organizationID,
		TargetStatus:   input.Body.TargetStatus,
		ReviewNote:     input.Body.ReviewNote,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.CooperationApplicationReviewResponse{Status: http.StatusOK}
	out.Body = *organizationReviewCooperationApplicationDTO(result)
	return out, nil
}

func (h *Handler) TransitionLegalDocumentReview(ctx context.Context, input *platformdto.TransitionLegalDocumentReviewInput) (*platformdto.LegalDocumentReviewResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documentID, err := httpbind.ParseUUID(input.DocumentID, fault.Validation("Document id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("documentId", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	result, err := h.svc.TransitionLegalDocumentReview(ctx, platformapp.TransitionLegalDocumentReviewCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		OrganizationID: organizationID,
		DocumentID:     documentID,
		TargetStatus:   input.Body.TargetStatus,
		ReviewNote:     input.Body.ReviewNote,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.LegalDocumentReviewResponse{Status: http.StatusOK}
	out.Body = organizationReviewLegalDocumentDTO(*result)
	return out, nil
}

func (h *Handler) ForceVerifyUserEmail(ctx context.Context, input *platformdto.ForceVerifyUserEmailInput) (*platformdto.ForceVerifyUserEmailResponse, error) {
	access := platformAccessFromContext(ctx)
	if access == nil {
		return nil, humaerr.From(ctx, fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN")))
	}
	res, err := h.svc.ForceVerifyUserEmail(ctx, platformapp.ForceVerifyUserEmailCmd{
		ActorAccountID: access.AccountID,
		ActorRoles:     access.EffectiveRoles,
		ActorBootstrap: access.BootstrapAdmin,
		UserID:         input.UserID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &platformdto.ForceVerifyUserEmailResponse{Status: http.StatusOK}
	out.Body.UserID = res.UserID
	out.Body.Email = res.Email
	out.Body.Verified = res.Verified
	out.Body.AlreadyVerified = res.AlreadyVerified
	return out, nil
}

func accessResponse(status int, access *platformdomain.Access) *platformdto.PlatformAccessResponse {
	out := &platformdto.PlatformAccessResponse{Status: status}
	out.Body.AccountID = access.AccountID
	out.Body.StoredRoles = platformdomain.RoleStrings(access.StoredRoles)
	out.Body.EffectiveRoles = platformdomain.RoleStrings(access.EffectiveRoles)
	out.Body.BootstrapAdmin = access.BootstrapAdmin
	return out
}

func autoGrantRuleDTO(rule platformdomain.AutoGrantRule) platformdto.AutoGrantRule {
	return platformdto.AutoGrantRule{
		ID:                 rule.ID,
		Role:               string(rule.Role),
		MatchType:          string(rule.MatchType),
		MatchValue:         rule.MatchValue,
		Source:             string(rule.Source),
		CreatedByAccountID: rule.CreatedByAccountID,
		CreatedAt:          rule.CreatedAt,
		UpdatedAt:          rule.UpdatedAt,
	}
}

func organizationReviewQueueItemDTO(item platformdomain.OrganizationReviewQueueItem) platformdto.OrganizationReviewQueueItem {
	return platformdto.OrganizationReviewQueueItem{
		OrganizationID:           item.OrganizationID,
		OrganizationName:         item.OrganizationName,
		OrganizationSlug:         item.OrganizationSlug,
		OrganizationIsActive:     item.OrganizationIsActive,
		CooperationApplicationID: item.CooperationApplicationID,
		CooperationStatus:        item.CooperationStatus,
		CompanyName:              item.CompanyName,
		ConfirmationEmail:        item.ConfirmationEmail,
		ReviewerAccountID:        item.ReviewerAccountID,
		SubmittedAt:              item.SubmittedAt,
		ReviewedAt:               item.ReviewedAt,
		CreatedAt:                item.CreatedAt,
		UpdatedAt:                item.UpdatedAt,
	}
}

func organizationReviewOrganizationDTO(item platformdomain.OrganizationReviewOrganization) platformdto.OrganizationReviewOrganization {
	return platformdto.OrganizationReviewOrganization{
		ID:           item.ID,
		Name:         item.Name,
		Slug:         item.Slug,
		LogoObjectID: item.LogoObjectID,
		Description:  item.Description,
		Website:      item.Website,
		PrimaryEmail: item.PrimaryEmail,
		Phone:        item.Phone,
		Address:      item.Address,
		Industry:     item.Industry,
		IsActive:     item.IsActive,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func organizationReviewDomainDTO(item platformdomain.OrganizationReviewDomain) platformdto.OrganizationReviewDomain {
	return platformdto.OrganizationReviewDomain{
		ID:         item.ID,
		Hostname:   item.Hostname,
		Kind:       item.Kind,
		IsPrimary:  item.IsPrimary,
		IsVerified: item.IsVerified,
		VerifiedAt: item.VerifiedAt,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}

func organizationReviewCooperationApplicationDTO(item *platformdomain.OrganizationReviewCooperationApplication) *platformdto.OrganizationReviewCooperationApplication {
	if item == nil {
		return nil
	}
	return &platformdto.OrganizationReviewCooperationApplication{
		ID:                    item.ID,
		OrganizationID:        item.OrganizationID,
		Status:                item.Status,
		ConfirmationEmail:     item.ConfirmationEmail,
		CompanyName:           item.CompanyName,
		RepresentedCategories: item.RepresentedCategories,
		MinimumOrderAmount:    item.MinimumOrderAmount,
		DeliveryGeography:     item.DeliveryGeography,
		SalesChannels:         append([]string{}, item.SalesChannels...),
		StorefrontURL:         item.StorefrontURL,
		ContactFirstName:      item.ContactFirstName,
		ContactLastName:       item.ContactLastName,
		ContactJobTitle:       item.ContactJobTitle,
		PriceListObjectID:     item.PriceListObjectID,
		ContactEmail:          item.ContactEmail,
		ContactPhone:          item.ContactPhone,
		PartnerCode:           item.PartnerCode,
		ReviewNote:            item.ReviewNote,
		ReviewerAccountID:     item.ReviewerAccountID,
		SubmittedAt:           item.SubmittedAt,
		ReviewedAt:            item.ReviewedAt,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}

func organizationReviewLegalDocumentDTO(item platformdomain.OrganizationReviewLegalDocument) platformdto.OrganizationReviewLegalDocument {
	out := platformdto.OrganizationReviewLegalDocument{
		ID:                  item.ID,
		OrganizationID:      item.OrganizationID,
		DocumentType:        item.DocumentType,
		Status:              item.Status,
		ObjectID:            item.ObjectID,
		Title:               item.Title,
		UploadedByAccountID: item.UploadedByAccountID,
		ReviewerAccountID:   item.ReviewerAccountID,
		ReviewNote:          item.ReviewNote,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
		ReviewedAt:          item.ReviewedAt,
	}
	if item.Analysis != nil {
		out.Analysis = &platformdto.OrganizationReviewLegalDocumentAnalysis{
			ID:                   item.Analysis.ID,
			DocumentID:           item.Analysis.DocumentID,
			OrganizationID:       item.Analysis.OrganizationID,
			Status:               item.Analysis.Status,
			Provider:             item.Analysis.Provider,
			Summary:              item.Analysis.Summary,
			DetectedDocumentType: item.Analysis.DetectedDocumentType,
			ConfidenceScore:      item.Analysis.ConfidenceScore,
			RequestedAt:          item.Analysis.RequestedAt,
			StartedAt:            item.Analysis.StartedAt,
			CompletedAt:          item.Analysis.CompletedAt,
			UpdatedAt:            item.Analysis.UpdatedAt,
			LastError:            item.Analysis.LastError,
		}
	}
	if item.Verification != nil {
		out.Verification = &platformdto.OrganizationReviewLegalDocumentVerification{
			DocumentID:           item.Verification.DocumentID,
			OrganizationID:       item.Verification.OrganizationID,
			DocumentType:         item.Verification.DocumentType,
			DocumentStatus:       item.Verification.DocumentStatus,
			AnalysisStatus:       item.Verification.AnalysisStatus,
			Verdict:              item.Verification.Verdict,
			Summary:              item.Verification.Summary,
			DetectedDocumentType: item.Verification.DetectedDocumentType,
			ConfidenceScore:      item.Verification.ConfidenceScore,
			RequiredFields:       append([]string{}, item.Verification.RequiredFields...),
			MissingFields:        append([]string{}, item.Verification.MissingFields...),
			CheckedAt:            item.Verification.CheckedAt,
		}
		out.Verification.Issues = make([]platformdto.OrganizationReviewLegalDocumentVerificationIssue, 0, len(item.Verification.Issues))
		for _, issue := range item.Verification.Issues {
			out.Verification.Issues = append(out.Verification.Issues, platformdto.OrganizationReviewLegalDocumentVerificationIssue{
				Code:     issue.Code,
				Severity: issue.Severity,
				Message:  issue.Message,
				Field:    issue.Field,
			})
		}
	}
	return out
}

func organizationReviewKYCDTO(item *platformdomain.OrganizationReviewKYCRequirements) *platformdto.OrganizationReviewKYCRequirements {
	if item == nil {
		return nil
	}
	out := &platformdto.OrganizationReviewKYCRequirements{
		OrganizationID: item.OrganizationID,
		Status:         item.Status,
		DisabledReason: item.DisabledReason,
		CheckedAt:      item.CheckedAt,
	}
	out.CurrentlyDue = organizationReviewKYCItemsDTO(item.CurrentlyDue)
	out.PendingVerification = organizationReviewKYCItemsDTO(item.PendingVerification)
	out.EventuallyDue = organizationReviewKYCItemsDTO(item.EventuallyDue)
	out.Errors = organizationReviewKYCItemsDTO(item.Errors)
	return out
}

func organizationReviewKYCItemsDTO(items []platformdomain.OrganizationReviewKYCRequirementItem) []platformdto.OrganizationReviewKYCRequirementItem {
	if len(items) == 0 {
		return []platformdto.OrganizationReviewKYCRequirementItem{}
	}
	out := make([]platformdto.OrganizationReviewKYCRequirementItem, 0, len(items))
	for _, item := range items {
		out = append(out, platformdto.OrganizationReviewKYCRequirementItem{
			Code:         item.Code,
			Category:     item.Category,
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

func parseOptionalUUID(raw string, field string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := httpbind.ParseUUID(raw, fault.Validation(fmt.Sprintf("%s is invalid", field), fault.Code("PLATFORM_INVALID_INPUT"), fault.Field(field, "must be a UUID")))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalString(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}
