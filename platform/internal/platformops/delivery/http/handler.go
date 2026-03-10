package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
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
