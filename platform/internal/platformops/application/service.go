package application

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/platformops/application/ports"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

type ReplaceAccountRolesCmd struct {
	ActorAccountID  uuid.UUID
	ActorRoles      []domain.Role
	ActorBootstrap  bool
	TargetAccountID uuid.UUID
	Roles           []string
}

type AddAutoGrantRuleCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	Role           string
	MatchType      string
	MatchValue     string
}

type DeleteAutoGrantRuleCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	RuleID         uuid.UUID
}

type ForceVerifyUserEmailCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	UserID         string
}

type ForceVerifyUserEmailResult struct {
	UserID          string
	Email           string
	Verified        bool
	AlreadyVerified bool
}

type AuditDeniedCmd struct {
	ActorAccountID *uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	Action         string
	TargetType     string
	TargetID       string
	Summary        string
}

type ListUploadQueueCmd struct {
	Status             *string
	Purpose            *string
	OrganizationID     *uuid.UUID
	CreatedByAccountID *uuid.UUID
	Limit              int
	Offset             int
}

type ListOrganizationReviewQueueCmd struct {
	Status            *string
	OrganizationID    *uuid.UUID
	ReviewerAccountID *uuid.UUID
	Search            *string
	Limit             int
	Offset            int
}

type TransitionCooperationApplicationReviewCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	OrganizationID uuid.UUID
	TargetStatus   string
	ReviewNote     *string
}

type TransitionLegalDocumentReviewCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	OrganizationID uuid.UUID
	DocumentID     uuid.UUID
	TargetStatus   string
	ReviewNote     *string
}

type Service struct {
	roles           ports.RoleBindingRepository
	autoGrants      ports.AutoGrantRuleRepository
	audits          ports.AuditRepository
	accounts        ports.AccountReader
	dashboards      ports.DashboardReader
	uploads         ports.UploadQueueReader
	reviews         ports.OrganizationReviewRepository
	clock           ports.Clock
	txm             sharedtx.Manager
	zitadelAdmin    authports.ZitadelAdminClient
	bootstrapAdmins map[uuid.UUID]struct{}
}

func New(
	roles ports.RoleBindingRepository,
	autoGrants ports.AutoGrantRuleRepository,
	audits ports.AuditRepository,
	accounts ports.AccountReader,
	dashboards ports.DashboardReader,
	uploads ports.UploadQueueReader,
	reviews ports.OrganizationReviewRepository,
	clock ports.Clock,
	txm sharedtx.Manager,
	zitadelAdmin authports.ZitadelAdminClient,
	bootstrapAccountIDs []uuid.UUID,
) *Service {
	bootstrapAdmins := make(map[uuid.UUID]struct{}, len(bootstrapAccountIDs))
	for _, accountID := range bootstrapAccountIDs {
		if accountID == uuid.Nil {
			continue
		}
		bootstrapAdmins[accountID] = struct{}{}
	}
	return &Service{
		roles:           roles,
		autoGrants:      autoGrants,
		audits:          audits,
		accounts:        accounts,
		dashboards:      dashboards,
		uploads:         uploads,
		reviews:         reviews,
		clock:           clock,
		txm:             txm,
		zitadelAdmin:    zitadelAdmin,
		bootstrapAdmins: bootstrapAdmins,
	}
}

func (s *Service) ZitadelAdminEnabled() bool {
	return s != nil && hasZitadelAdminClient(s.zitadelAdmin)
}

func (s *Service) ResolveAccess(ctx context.Context, accountID uuid.UUID) (*domain.Access, error) {
	if accountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	storedRoles, err := s.roles.ListRoles(ctx, accountID)
	if err != nil {
		return nil, fault.Internal("Resolve platform access failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	return s.buildAccess(accountID, storedRoles), nil
}

func (s *Service) GetAccountAccess(ctx context.Context, accountID uuid.UUID) (*domain.Access, error) {
	targetID, err := accdomain.AccountIDFromUUID(accountID)
	if err != nil {
		return nil, fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID"))
	}
	account, err := s.accounts.GetByID(ctx, targetID)
	if err != nil {
		return nil, fault.Internal("Load account failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	if account == nil {
		return nil, fault.NotFound("Account not found", fault.Code("PLATFORM_ACCOUNT_NOT_FOUND"))
	}
	return s.ResolveAccess(ctx, accountID)
}

func (s *Service) ReplaceAccountRoles(ctx context.Context, cmd ReplaceAccountRolesCmd) (*domain.Access, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	targetID, err := accdomain.AccountIDFromUUID(cmd.TargetAccountID)
	if err != nil {
		err = fault.Validation("Account id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("accountId", "must be a UUID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}
	account, err := s.accounts.GetByID(ctx, targetID)
	if err != nil {
		err = fault.Internal("Load account failed", fault.Code("INTERNAL"), fault.WithCause(err))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}
	if account == nil {
		err = fault.NotFound("Account not found", fault.Code("PLATFORM_ACCOUNT_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}

	roles, err := normalizeInputRoles(cmd.Roles)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}

	currentAdminIDs, err := s.roles.ListAccountIDsByRole(ctx, domain.RolePlatformAdmin)
	if err != nil {
		err = fault.Internal("Load platform admin bindings failed", fault.Code("INTERNAL"), fault.WithCause(err))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}
	if !s.hasAnyAdminAfterChange(currentAdminIDs, cmd.TargetAccountID, roles) {
		err = fault.Conflict("Platform must retain at least one platform admin", fault.Code("PLATFORM_LAST_ADMIN"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}

	now := s.clock.Now()
	actorID := cmd.ActorAccountID
	if err := s.txm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.roles.ReplaceRoles(txCtx, cmd.TargetAccountID, roles, &actorID, now); err != nil {
			return fault.Internal("Update platform roles failed", fault.Code("INTERNAL"), fault.WithCause(err))
		}
		summary := summarizeRoles("storedRoles", roles)
		return s.audits.Append(txCtx, domain.AuditEvent{
			ActorAccountID: &actorID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.access.roles.replace",
			TargetType:     "account",
			TargetID:       stringPtr(cmd.TargetAccountID.String()),
			Status:         domain.AuditStatusSuccess,
			Summary:        stringPtr(summary),
			CreatedAt:      now,
		})
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.roles.replace", "account", cmd.TargetAccountID.String(), err.Error())
		return nil, err
	}

	return s.ResolveAccess(ctx, cmd.TargetAccountID)
}

func (s *Service) ListAutoGrantRules(ctx context.Context) ([]domain.AutoGrantRule, error) {
	rules, err := s.autoGrants.ListAutoGrantRules(ctx)
	if err != nil {
		return nil, fault.Internal("Load platform auto-grant rules failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	return rules, nil
}

func (s *Service) AddAutoGrantRule(ctx context.Context, cmd AddAutoGrantRuleCmd) (*domain.AutoGrantRule, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	role, matchType, matchValue, err := normalizeAutoGrantRuleInput(cmd.Role, cmd.MatchType, cmd.MatchValue)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.create", "platform_auto_grant_rule", "", err.Error())
		return nil, err
	}
	existingRules, err := s.autoGrants.ListAutoGrantRules(ctx)
	if err != nil {
		err = fault.Internal("Load platform auto-grant rules failed", fault.Code("INTERNAL"), fault.WithCause(err))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.create", "platform_auto_grant_rule", "", err.Error())
		return nil, err
	}
	for _, existing := range existingRules {
		if existing.Role != role || existing.MatchType != matchType || existing.MatchValue != matchValue {
			continue
		}
		message := "Platform auto-grant rule already exists"
		if existing.Source == domain.AutoGrantSourceBootstrap {
			message = "Platform auto-grant rule already exists in bootstrap config"
		}
		conflictErr := fault.Conflict(message, fault.Code("PLATFORM_AUTO_GRANT_EXISTS"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.create", "platform_auto_grant_rule", derefUUID(existing.ID), conflictErr.Error())
		return nil, conflictErr
	}
	actorID := cmd.ActorAccountID
	now := s.clock.Now()
	var created *domain.AutoGrantRule
	if err := s.txm.WithinTransaction(ctx, func(txCtx context.Context) error {
		created, err = s.autoGrants.CreateAutoGrantRule(txCtx, role, matchType, matchValue, &actorID, now)
		if err != nil {
			return fault.Internal("Create platform auto-grant rule failed", fault.Code("INTERNAL"), fault.WithCause(err))
		}
		summary := fmt.Sprintf("role=%s matchType=%s matchValue=%s", created.Role, created.MatchType, created.MatchValue)
		return s.audits.Append(txCtx, domain.AuditEvent{
			ActorAccountID: &actorID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.access.auto_grant_rule.create",
			TargetType:     "platform_auto_grant_rule",
			TargetID:       stringPtr(derefUUID(created.ID)),
			Status:         domain.AuditStatusSuccess,
			Summary:        stringPtr(summary),
			CreatedAt:      now,
		})
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.create", "platform_auto_grant_rule", derefUUID(createdID(created)), err.Error())
		return nil, err
	}
	return created, nil
}

func (s *Service) DeleteAutoGrantRule(ctx context.Context, cmd DeleteAutoGrantRuleCmd) (*domain.AutoGrantRule, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	if cmd.RuleID == uuid.Nil {
		err := fault.Validation("Auto-grant rule id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("ruleId", "must be a UUID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.delete", "platform_auto_grant_rule", "", err.Error())
		return nil, err
	}
	actorID := cmd.ActorAccountID
	now := s.clock.Now()
	var deleted *domain.AutoGrantRule
	if err := s.txm.WithinTransaction(ctx, func(txCtx context.Context) error {
		deletedRule, opErr := s.autoGrants.DeleteAutoGrantRule(txCtx, cmd.RuleID)
		if opErr != nil {
			return fault.Internal("Delete platform auto-grant rule failed", fault.Code("INTERNAL"), fault.WithCause(opErr))
		}
		deleted = deletedRule
		if deleted == nil {
			return fault.NotFound("Platform auto-grant rule not found", fault.Code("PLATFORM_AUTO_GRANT_NOT_FOUND"))
		}
		summary := fmt.Sprintf("role=%s matchType=%s matchValue=%s", deleted.Role, deleted.MatchType, deleted.MatchValue)
		return s.audits.Append(txCtx, domain.AuditEvent{
			ActorAccountID: &actorID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.access.auto_grant_rule.delete",
			TargetType:     "platform_auto_grant_rule",
			TargetID:       stringPtr(cmd.RuleID.String()),
			Status:         domain.AuditStatusSuccess,
			Summary:        stringPtr(summary),
			CreatedAt:      now,
		})
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.access.auto_grant_rule.delete", "platform_auto_grant_rule", cmd.RuleID.String(), err.Error())
		return nil, err
	}
	return deleted, nil
}

func (s *Service) ForceVerifyUserEmail(ctx context.Context, cmd ForceVerifyUserEmailCmd) (*ForceVerifyUserEmailResult, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		err := fault.Validation("ZITADEL user id is required", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("userId", "is required"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.user.email.force_verify", "zitadel_user", userID, err.Error())
		return nil, err
	}
	if !hasZitadelAdminClient(s.zitadelAdmin) {
		err := fault.Forbidden("ZITADEL admin email verification is disabled. Configure AUTH_ZITADEL_ADMIN_TOKEN to enable it.", fault.Code("PLATFORM_ZITADEL_ADMIN_DISABLED"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.user.email.force_verify", "zitadel_user", userID, err.Error())
		return nil, err
	}

	res, err := s.zitadelAdmin.ForceVerifyUserEmail(ctx, userID)
	if err != nil {
		mapped := mapZitadelAdminError(err)
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.user.email.force_verify", "zitadel_user", userID, mapped.Error())
		return nil, mapped
	}

	now := s.clock.Now()
	actorID := cmd.ActorAccountID
	summary := nonEmpty(
		fmt.Sprintf("email=%s", strings.TrimSpace(res.Email)),
		fmt.Sprintf("alreadyVerified=%t", res.AlreadyVerified),
	)
	s.appendAuditBestEffort(ctx, domain.AuditEvent{
		ActorAccountID: &actorID,
		ActorRoles:     cmd.ActorRoles,
		ActorBootstrap: cmd.ActorBootstrap,
		Action:         "platform.user.email.force_verify",
		TargetType:     "zitadel_user",
		TargetID:       stringPtr(userID),
		Status:         domain.AuditStatusSuccess,
		Summary:        stringPtr(summary),
		CreatedAt:      now,
	})

	return &ForceVerifyUserEmailResult{
		UserID:          res.UserID,
		Email:           res.Email,
		Verified:        true,
		AlreadyVerified: res.AlreadyVerified,
	}, nil
}

func (s *Service) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	summary, err := s.dashboards.GetDashboardSummary(ctx)
	if err != nil {
		return nil, fault.Internal("Load platform dashboard summary failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	return summary, nil
}

func (s *Service) ListUploadQueue(ctx context.Context, cmd ListUploadQueueCmd) ([]domain.UploadQueueItem, int, error) {
	query := domain.UploadQueueQuery{
		Status:             normalizeOptionalUploadStatus(cmd.Status),
		Purpose:            normalizeOptionalUploadPurpose(cmd.Purpose),
		OrganizationID:     cmd.OrganizationID,
		CreatedByAccountID: cmd.CreatedByAccountID,
		Limit:              cmd.Limit,
		Offset:             cmd.Offset,
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if err := validateUploadQuery(query); err != nil {
		return nil, 0, err
	}
	items, total, err := s.uploads.ListUploadQueue(ctx, query)
	if err != nil {
		return nil, 0, fault.Internal("Load upload queue failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	return items, total, nil
}

func (s *Service) ListOrganizationReviewQueue(ctx context.Context, cmd ListOrganizationReviewQueueCmd) ([]domain.OrganizationReviewQueueItem, int, error) {
	query := domain.OrganizationReviewQueueQuery{
		Status:            normalizeOptionalOrganizationReviewStatus(cmd.Status),
		OrganizationID:    cmd.OrganizationID,
		ReviewerAccountID: cmd.ReviewerAccountID,
		Search:            normalizeOptionalSearch(cmd.Search),
		Limit:             cmd.Limit,
		Offset:            cmd.Offset,
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if err := validateOrganizationReviewQueueQuery(query); err != nil {
		return nil, 0, err
	}
	items, total, err := s.reviews.ListOrganizationReviewQueue(ctx, query)
	if err != nil {
		return nil, 0, fault.Internal("Load organization review queue failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	return items, total, nil
}

func (s *Service) GetOrganizationReview(ctx context.Context, organizationID uuid.UUID) (*domain.OrganizationReviewDetail, error) {
	if organizationID == uuid.Nil {
		return nil, fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID"))
	}
	detail, err := s.reviews.GetOrganizationReview(ctx, organizationID)
	if err != nil {
		return nil, fault.Internal("Load organization review failed", fault.Code("INTERNAL"), fault.WithCause(err))
	}
	if detail == nil {
		return nil, fault.NotFound("Organization review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
	}
	enrichOrganizationReviewDetailWithVerification(detail, s.clock.Now())
	enrichOrganizationReviewDetailWithKYC(detail, s.clock.Now())
	return detail, nil
}

func (s *Service) TransitionCooperationApplicationReview(ctx context.Context, cmd TransitionCooperationApplicationReviewCmd) (*domain.OrganizationReviewCooperationApplication, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	if cmd.OrganizationID == uuid.Nil {
		err := fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", "", err.Error())
		return nil, err
	}
	if !canTransitionOrganizationReviews(cmd.ActorRoles) {
		err := fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN"))
		s.RecordDeniedAudit(ctx, AuditDeniedCmd{
			ActorAccountID: &cmd.ActorAccountID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.organization.review.transition",
			TargetType:     "organization",
			TargetID:       cmd.OrganizationID.String(),
			Summary:        "required roles: platform_admin, review_operator",
		})
		return nil, err
	}

	targetStatus, err := normalizeCooperationApplicationTransitionTarget(cmd.TargetStatus)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}
	reviewNote, err := normalizeReviewNote(cmd.ReviewNote)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}
	if requiresReviewNote(targetStatus) && reviewNote == nil {
		err = fault.Validation("Review note is required for this transition", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("reviewNote", "is required"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}

	detail, err := s.reviews.GetOrganizationReview(ctx, cmd.OrganizationID)
	if err != nil {
		err = fault.Internal("Load organization review failed", fault.Code("INTERNAL"), fault.WithCause(err))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}
	if detail == nil || detail.CooperationApplication == nil {
		err = fault.NotFound("Cooperation application review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}

	currentStatus := normalizeOptionalOrganizationReviewStatus(stringPtr(detail.CooperationApplication.Status))
	if !isAllowedCooperationApplicationTransition(derefString(currentStatus), targetStatus) {
		err = fault.Conflict("Cooperation application transition is not allowed", fault.Code("PLATFORM_REVIEW_TRANSITION_INVALID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}

	now := s.clock.Now()
	actorID := cmd.ActorAccountID
	patch := domain.CooperationApplicationReviewPatch{
		Status:            targetStatus,
		ReviewNote:        reviewNote,
		ReviewerAccountID: &actorID,
		UpdatedAt:         now,
	}
	if targetStatus != string(orgdomain.CooperationApplicationStatusUnderReview) {
		patch.ReviewedAt = &now
	}

	var updated *domain.OrganizationReviewCooperationApplication
	if err := s.txm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var opErr error
		updated, opErr = s.reviews.UpdateCooperationApplicationReview(txCtx, cmd.OrganizationID, patch)
		if opErr != nil {
			return fault.Internal("Update cooperation application review failed", fault.Code("INTERNAL"), fault.WithCause(opErr))
		}
		if updated == nil {
			return fault.NotFound("Cooperation application review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
		}
		summary := fmt.Sprintf("from=%s to=%s", detail.CooperationApplication.Status, updated.Status)
		return s.audits.Append(txCtx, domain.AuditEvent{
			ActorAccountID: &actorID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.organization.review.transition",
			TargetType:     "organization",
			TargetID:       stringPtr(cmd.OrganizationID.String()),
			Status:         domain.AuditStatusSuccess,
			Summary:        stringPtr(summary),
			CreatedAt:      now,
		})
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.review.transition", "organization", cmd.OrganizationID.String(), err.Error())
		return nil, err
	}

	return updated, nil
}

func (s *Service) TransitionLegalDocumentReview(ctx context.Context, cmd TransitionLegalDocumentReviewCmd) (*domain.OrganizationReviewLegalDocument, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	if cmd.OrganizationID == uuid.Nil {
		err := fault.Validation("Organization id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("organizationId", "must be a UUID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", "", err.Error())
		return nil, err
	}
	if cmd.DocumentID == uuid.Nil {
		err := fault.Validation("Document id is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("documentId", "must be a UUID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", "", err.Error())
		return nil, err
	}
	if !canTransitionOrganizationReviews(cmd.ActorRoles) {
		err := fault.Forbidden("Platform access denied", fault.Code("PLATFORM_FORBIDDEN"))
		s.RecordDeniedAudit(ctx, AuditDeniedCmd{
			ActorAccountID: &cmd.ActorAccountID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.organization.legal_document.review.transition",
			TargetType:     "organization_legal_document",
			TargetID:       cmd.DocumentID.String(),
			Summary:        "required roles: platform_admin, review_operator",
		})
		return nil, err
	}

	targetStatus, err := normalizeLegalDocumentTransitionTarget(cmd.TargetStatus)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}
	reviewNote, err := normalizeReviewNote(cmd.ReviewNote)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}
	if requiresLegalDocumentReviewNote(targetStatus) && reviewNote == nil {
		err = fault.Validation("Review note is required for this transition", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("reviewNote", "is required"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}

	detail, err := s.reviews.GetOrganizationReview(ctx, cmd.OrganizationID)
	if err != nil {
		err = fault.Internal("Load organization review failed", fault.Code("INTERNAL"), fault.WithCause(err))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}
	if detail == nil {
		err = fault.NotFound("Organization review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}

	current := findOrganizationReviewLegalDocument(detail.LegalDocuments, cmd.DocumentID)
	if current == nil {
		err = fault.NotFound("Legal document review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}
	if !isAllowedLegalDocumentTransition(current.Status, targetStatus) {
		err = fault.Conflict("Legal document transition is not allowed", fault.Code("PLATFORM_REVIEW_TRANSITION_INVALID"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}

	now := s.clock.Now()
	actorID := cmd.ActorAccountID
	patch := domain.LegalDocumentReviewPatch{
		Status:            targetStatus,
		ReviewNote:        reviewNote,
		ReviewerAccountID: &actorID,
		ReviewedAt:        &now,
		UpdatedAt:         now,
	}

	var updated *domain.OrganizationReviewLegalDocument
	if err := s.txm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var opErr error
		updated, opErr = s.reviews.UpdateLegalDocumentReview(txCtx, cmd.OrganizationID, cmd.DocumentID, patch)
		if opErr != nil {
			return fault.Internal("Update legal document review failed", fault.Code("INTERNAL"), fault.WithCause(opErr))
		}
		if updated == nil {
			return fault.NotFound("Legal document review not found", fault.Code("PLATFORM_REVIEW_NOT_FOUND"))
		}
		summary := fmt.Sprintf("from=%s to=%s", current.Status, updated.Status)
		return s.audits.Append(txCtx, domain.AuditEvent{
			ActorAccountID: &actorID,
			ActorRoles:     cmd.ActorRoles,
			ActorBootstrap: cmd.ActorBootstrap,
			Action:         "platform.organization.legal_document.review.transition",
			TargetType:     "organization_legal_document",
			TargetID:       stringPtr(cmd.DocumentID.String()),
			Status:         domain.AuditStatusSuccess,
			Summary:        stringPtr(summary),
			CreatedAt:      now,
		})
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.organization.legal_document.review.transition", "organization_legal_document", cmd.DocumentID.String(), err.Error())
		return nil, err
	}

	enrichOrganizationReviewLegalDocumentWithVerification(updated, s.clock.Now())
	return updated, nil
}

func (s *Service) RecordDeniedAudit(ctx context.Context, cmd AuditDeniedCmd) {
	var targetID *string
	if strings.TrimSpace(cmd.TargetID) != "" {
		targetID = stringPtr(cmd.TargetID)
	}
	s.appendAuditBestEffort(ctx, domain.AuditEvent{
		ActorAccountID: cmd.ActorAccountID,
		ActorRoles:     cmd.ActorRoles,
		ActorBootstrap: cmd.ActorBootstrap,
		Action:         strings.TrimSpace(cmd.Action),
		TargetType:     nonEmpty(cmd.TargetType, "operation"),
		TargetID:       targetID,
		Status:         domain.AuditStatusDenied,
		Summary:        stringPtr(strings.TrimSpace(cmd.Summary)),
		CreatedAt:      s.clock.Now(),
	})
}

func (s *Service) buildAccess(accountID uuid.UUID, storedRoles []domain.Role) *domain.Access {
	stored := domain.UniqueSortedRoles(storedRoles)
	effective := append([]domain.Role{}, stored...)
	bootstrap := false
	if _, ok := s.bootstrapAdmins[accountID]; ok {
		bootstrap = true
		effective = append(effective, domain.RolePlatformAdmin)
	}
	effective = domain.UniqueSortedRoles(effective)
	return &domain.Access{
		AccountID:      accountID,
		StoredRoles:    stored,
		EffectiveRoles: effective,
		BootstrapAdmin: bootstrap,
	}
}

func (s *Service) hasAnyAdminAfterChange(currentAdminIDs []uuid.UUID, targetAccountID uuid.UUID, newRoles []domain.Role) bool {
	admins := make(map[uuid.UUID]struct{}, len(currentAdminIDs)+len(s.bootstrapAdmins)+1)
	for _, accountID := range currentAdminIDs {
		if accountID == targetAccountID {
			continue
		}
		admins[accountID] = struct{}{}
	}
	for accountID := range s.bootstrapAdmins {
		admins[accountID] = struct{}{}
	}
	if containsRole(newRoles, domain.RolePlatformAdmin) {
		admins[targetAccountID] = struct{}{}
	}
	return len(admins) > 0
}

func normalizeInputRoles(raw []string) ([]domain.Role, error) {
	parsed := make([]domain.Role, 0, len(raw))
	seenInvalid := make([]string, 0)
	for _, item := range raw {
		role := domain.ParseRole(item)
		if !role.IsValid() {
			if strings.TrimSpace(item) != "" {
				seenInvalid = append(seenInvalid, strings.TrimSpace(item))
			}
			continue
		}
		parsed = append(parsed, role)
	}
	if len(seenInvalid) > 0 {
		return nil, fault.Validation(
			fmt.Sprintf("Unsupported platform roles: %s", strings.Join(seenInvalid, ", ")),
			fault.Code("PLATFORM_INVALID_INPUT"),
			fault.Field("roles", "contains unsupported values"),
		)
	}
	return domain.UniqueSortedRoles(parsed), nil
}

func normalizeAutoGrantRuleInput(rawRole string, rawMatchType string, rawMatchValue string) (domain.Role, domain.AutoGrantMatchType, string, error) {
	role := domain.ParseRole(rawRole)
	if !role.IsValid() {
		return "", "", "", fault.Validation("Unsupported platform role", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("role", "must be platform_admin, support_operator, or review_operator"))
	}
	matchType := domain.ParseAutoGrantMatchType(rawMatchType)
	if !matchType.IsValid() {
		return "", "", "", fault.Validation("Unsupported auto-grant match type", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("matchType", "must be email or subject"))
	}
	matchValue := domain.NormalizeAutoGrantMatchValue(matchType, rawMatchValue)
	if matchValue == "" {
		return "", "", "", fault.Validation("Auto-grant match value is required", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("matchValue", "is required"))
	}
	return role, matchType, matchValue, nil
}

func normalizeOptionalUploadStatus(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalUploadPurpose(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalSearch(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateUploadQuery(query domain.UploadQueueQuery) error {
	if query.Status != nil {
		switch *query.Status {
		case string(uploaddomain.StatusPending), string(uploaddomain.StatusReady), string(uploaddomain.StatusFailed):
		default:
			return fault.Validation("Unsupported upload status", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("status", "must be pending, ready, or failed"))
		}
	}
	if query.Purpose != nil {
		switch *query.Purpose {
		case string(uploaddomain.PurposeOrganizationLegalDocument), string(uploaddomain.PurposeProductImport):
		default:
			return fault.Validation("Unsupported upload purpose", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("purpose", "must be organization_legal_document or product_import"))
		}
	}
	return nil
}

func validateOrganizationReviewQueueQuery(query domain.OrganizationReviewQueueQuery) error {
	if query.Status != nil {
		switch *query.Status {
		case string(orgdomain.CooperationApplicationStatusDraft),
			string(orgdomain.CooperationApplicationStatusSubmitted),
			string(orgdomain.CooperationApplicationStatusUnderReview),
			string(orgdomain.CooperationApplicationStatusApproved),
			string(orgdomain.CooperationApplicationStatusRejected),
			string(orgdomain.CooperationApplicationStatusNeedsInfo):
		default:
			return fault.Validation("Unsupported cooperation review status", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("status", "must be draft, submitted, under_review, approved, rejected, or needs_info"))
		}
	}
	return nil
}

func normalizeOptionalOrganizationReviewStatus(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeCooperationApplicationTransitionTarget(raw string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(raw))
	switch status {
	case string(orgdomain.CooperationApplicationStatusUnderReview),
		string(orgdomain.CooperationApplicationStatusApproved),
		string(orgdomain.CooperationApplicationStatusRejected),
		string(orgdomain.CooperationApplicationStatusNeedsInfo):
		return status, nil
	default:
		return "", fault.Validation("Unsupported cooperation review target status", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("targetStatus", "must be under_review, approved, rejected, or needs_info"))
	}
}

func normalizeReviewNote(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > 4096 {
		return nil, fault.Validation("Review note is too long", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("reviewNote", "must be 4096 characters or fewer"))
	}
	return &trimmed, nil
}

func requiresReviewNote(status string) bool {
	switch status {
	case string(orgdomain.CooperationApplicationStatusRejected), string(orgdomain.CooperationApplicationStatusNeedsInfo):
		return true
	default:
		return false
	}
}

func isAllowedCooperationApplicationTransition(from string, to string) bool {
	switch from {
	case string(orgdomain.CooperationApplicationStatusSubmitted):
		return to == string(orgdomain.CooperationApplicationStatusUnderReview)
	case string(orgdomain.CooperationApplicationStatusUnderReview):
		return to == string(orgdomain.CooperationApplicationStatusApproved) ||
			to == string(orgdomain.CooperationApplicationStatusRejected) ||
			to == string(orgdomain.CooperationApplicationStatusNeedsInfo)
	case string(orgdomain.CooperationApplicationStatusNeedsInfo):
		return to == string(orgdomain.CooperationApplicationStatusUnderReview)
	default:
		return false
	}
}

func normalizeLegalDocumentTransitionTarget(raw string) (string, error) {
	switch strings.TrimSpace(raw) {
	case string(orgdomain.OrganizationLegalDocumentStatusApproved):
		return string(orgdomain.OrganizationLegalDocumentStatusApproved), nil
	case string(orgdomain.OrganizationLegalDocumentStatusRejected):
		return string(orgdomain.OrganizationLegalDocumentStatusRejected), nil
	default:
		return "", fault.Validation("Unsupported legal document target status", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("targetStatus", "must be approved or rejected"))
	}
}

func requiresLegalDocumentReviewNote(status string) bool {
	return status == string(orgdomain.OrganizationLegalDocumentStatusRejected)
}

func isAllowedLegalDocumentTransition(from string, to string) bool {
	from = strings.TrimSpace(from)
	if from == "" || from == to {
		return false
	}
	switch from {
	case string(orgdomain.OrganizationLegalDocumentStatusPending),
		string(orgdomain.OrganizationLegalDocumentStatusApproved),
		string(orgdomain.OrganizationLegalDocumentStatusRejected):
		return to == string(orgdomain.OrganizationLegalDocumentStatusApproved) ||
			to == string(orgdomain.OrganizationLegalDocumentStatusRejected)
	default:
		return false
	}
}

func findOrganizationReviewLegalDocument(items []domain.OrganizationReviewLegalDocument, documentID uuid.UUID) *domain.OrganizationReviewLegalDocument {
	for idx := range items {
		if items[idx].ID == documentID {
			return &items[idx]
		}
	}
	return nil
}

func enrichOrganizationReviewDetailWithVerification(detail *domain.OrganizationReviewDetail, now time.Time) {
	if detail == nil {
		return
	}
	for idx := range detail.LegalDocuments {
		enrichOrganizationReviewLegalDocumentWithVerification(&detail.LegalDocuments[idx], now)
	}
}

func enrichOrganizationReviewDetailWithKYC(detail *domain.OrganizationReviewDetail, now time.Time) {
	if detail == nil {
		return
	}
	detail.KYC = mapOrganizationReviewKYCRequirements(orgdomain.BuildOrganizationKYCRequirements(orgdomain.OrganizationKYCRequirementsInput{
		OrganizationID:         detail.Organization.ID,
		CooperationApplication: toOrganizationKYCCooperationInput(detail.CooperationApplication),
		LegalDocuments:         toOrganizationKYCLegalDocumentInputs(detail.LegalDocuments),
		Now:                    now,
	}))
}

func enrichOrganizationReviewLegalDocumentWithVerification(item *domain.OrganizationReviewLegalDocument, now time.Time) {
	if item == nil {
		return
	}
	verification := orgdomain.BuildOrganizationLegalDocumentVerification(orgdomain.OrganizationLegalDocumentVerificationInput{
		Document: orgdomain.OrganizationLegalDocumentVerificationDocumentInput{
			ID:             item.ID,
			OrganizationID: item.OrganizationID,
			DocumentType:   item.DocumentType,
			DocumentStatus: item.Status,
		},
		Analysis: toOrganizationLegalDocumentVerificationAnalysisInput(item.Analysis),
		Now:      now,
	})
	item.Verification = mapOrganizationReviewLegalDocumentVerification(verification)
}

func toOrganizationLegalDocumentVerificationAnalysisInput(item *domain.OrganizationReviewLegalDocumentAnalysis) *orgdomain.OrganizationLegalDocumentVerificationAnalysisInput {
	if item == nil {
		return nil
	}
	return &orgdomain.OrganizationLegalDocumentVerificationAnalysisInput{
		Status:               item.Status,
		ExtractedFieldsJSON:  item.ExtractedFieldsJSON,
		DetectedDocumentType: item.DetectedDocumentType,
		ConfidenceScore:      item.ConfidenceScore,
		RequestedAt:          item.RequestedAt,
		CompletedAt:          item.CompletedAt,
		UpdatedAt:            item.UpdatedAt,
		LastError:            item.LastError,
	}
}

func mapOrganizationReviewLegalDocumentVerification(item *orgdomain.OrganizationLegalDocumentVerification) *domain.OrganizationReviewLegalDocumentVerification {
	if item == nil {
		return nil
	}
	out := &domain.OrganizationReviewLegalDocumentVerification{
		DocumentID:           item.DocumentID,
		OrganizationID:       item.OrganizationID,
		DocumentType:         item.DocumentType,
		DocumentStatus:       item.DocumentStatus,
		AnalysisStatus:       item.AnalysisStatus,
		Verdict:              string(item.Verdict),
		Summary:              item.Summary,
		DetectedDocumentType: item.DetectedDocumentType,
		ConfidenceScore:      item.ConfidenceScore,
		RequiredFields:       append([]string{}, item.RequiredFields...),
		MissingFields:        append([]string{}, item.MissingFields...),
		CheckedAt:            item.CheckedAt,
	}
	out.Issues = make([]domain.OrganizationReviewLegalDocumentVerificationIssue, 0, len(item.Issues))
	for _, issue := range item.Issues {
		out.Issues = append(out.Issues, domain.OrganizationReviewLegalDocumentVerificationIssue{
			Code:     issue.Code,
			Severity: string(issue.Severity),
			Message:  issue.Message,
			Field:    issue.Field,
		})
	}
	return out
}

func mapOrganizationReviewKYCRequirements(item *orgdomain.OrganizationKYCRequirements) *domain.OrganizationReviewKYCRequirements {
	if item == nil {
		return nil
	}
	out := &domain.OrganizationReviewKYCRequirements{
		OrganizationID: item.OrganizationID,
		Status:         string(item.Status),
		DisabledReason: item.DisabledReason,
		CheckedAt:      item.CheckedAt,
	}
	out.CurrentlyDue = mapOrganizationReviewKYCRequirementItems(item.CurrentlyDue)
	out.PendingVerification = mapOrganizationReviewKYCRequirementItems(item.PendingVerification)
	out.EventuallyDue = mapOrganizationReviewKYCRequirementItems(item.EventuallyDue)
	out.Errors = mapOrganizationReviewKYCRequirementItems(item.Errors)
	return out
}

func mapOrganizationReviewKYCRequirementItems(items []orgdomain.OrganizationKYCRequirementItem) []domain.OrganizationReviewKYCRequirementItem {
	if len(items) == 0 {
		return []domain.OrganizationReviewKYCRequirementItem{}
	}
	out := make([]domain.OrganizationReviewKYCRequirementItem, 0, len(items))
	for _, item := range items {
		out = append(out, domain.OrganizationReviewKYCRequirementItem{
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

func toOrganizationKYCCooperationInput(item *domain.OrganizationReviewCooperationApplication) *orgdomain.OrganizationKYCCooperationApplicationInput {
	if item == nil {
		return nil
	}
	return &orgdomain.OrganizationKYCCooperationApplicationInput{
		Status:                item.Status,
		ReviewNote:            item.ReviewNote,
		ConfirmationEmail:     item.ConfirmationEmail,
		CompanyName:           item.CompanyName,
		RepresentedCategories: item.RepresentedCategories,
		MinimumOrderAmount:    item.MinimumOrderAmount,
		DeliveryGeography:     item.DeliveryGeography,
		SalesChannels:         append([]string{}, item.SalesChannels...),
		PriceListObjectID:     item.PriceListObjectID,
		ContactFirstName:      item.ContactFirstName,
		ContactLastName:       item.ContactLastName,
		ContactJobTitle:       item.ContactJobTitle,
		ContactEmail:          item.ContactEmail,
		ContactPhone:          item.ContactPhone,
	}
}

func toOrganizationKYCLegalDocumentInputs(items []domain.OrganizationReviewLegalDocument) []orgdomain.OrganizationKYCLegalDocumentInput {
	if len(items) == 0 {
		return nil
	}
	out := make([]orgdomain.OrganizationKYCLegalDocumentInput, 0, len(items))
	for _, item := range items {
		out = append(out, orgdomain.OrganizationKYCLegalDocumentInput{
			ID:           item.ID,
			DocumentType: item.DocumentType,
			Status:       item.Status,
			ReviewNote:   item.ReviewNote,
			Verification: toOrganizationLegalDocumentVerification(item.Verification),
		})
	}
	return out
}

func toOrganizationLegalDocumentVerification(item *domain.OrganizationReviewLegalDocumentVerification) *orgdomain.OrganizationLegalDocumentVerification {
	if item == nil {
		return nil
	}
	out := &orgdomain.OrganizationLegalDocumentVerification{
		DocumentID:           item.DocumentID,
		OrganizationID:       item.OrganizationID,
		DocumentType:         item.DocumentType,
		DocumentStatus:       item.DocumentStatus,
		AnalysisStatus:       item.AnalysisStatus,
		Verdict:              orgdomain.OrganizationLegalDocumentVerificationVerdict(item.Verdict),
		Summary:              item.Summary,
		DetectedDocumentType: item.DetectedDocumentType,
		ConfidenceScore:      item.ConfidenceScore,
		RequiredFields:       append([]string{}, item.RequiredFields...),
		MissingFields:        append([]string{}, item.MissingFields...),
		CheckedAt:            item.CheckedAt,
	}
	out.Issues = make([]orgdomain.OrganizationLegalDocumentVerificationIssue, 0, len(item.Issues))
	for _, issue := range item.Issues {
		out.Issues = append(out.Issues, orgdomain.OrganizationLegalDocumentVerificationIssue{
			Code:     issue.Code,
			Severity: orgdomain.OrganizationLegalDocumentVerificationIssueSeverity(issue.Severity),
			Message:  issue.Message,
			Field:    issue.Field,
		})
	}
	return out
}

func (s *Service) recordFailure(ctx context.Context, actorAccountID uuid.UUID, actorRoles []domain.Role, actorBootstrap bool, action, targetType, targetID, summary string) {
	actorID := actorAccountID
	s.appendAuditBestEffort(ctx, domain.AuditEvent{
		ActorAccountID: &actorID,
		ActorRoles:     actorRoles,
		ActorBootstrap: actorBootstrap,
		Action:         action,
		TargetType:     targetType,
		TargetID:       stringPtr(targetID),
		Status:         domain.AuditStatusFailed,
		Summary:        stringPtr(summary),
		CreatedAt:      s.clock.Now(),
	})
}

func (s *Service) appendAuditBestEffort(ctx context.Context, event domain.AuditEvent) {
	if s == nil || s.audits == nil {
		return
	}
	if strings.TrimSpace(event.Action) == "" {
		return
	}
	if err := s.audits.Append(ctx, event); err != nil {
		slog.Default().Error("platform audit append failed",
			"event", "platform.audit.append_failed",
			"action", event.Action,
			"target_type", event.TargetType,
			"target_id", derefString(event.TargetID),
			"status", event.Status,
			"error", err,
		)
	}
}

func hasZitadelAdminClient(client authports.ZitadelAdminClient) bool {
	if client == nil {
		return false
	}
	value := reflect.ValueOf(client)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func mapZitadelAdminError(err error) error {
	var apiErr *authports.ZitadelAdminAPIError
	if stderrors.As(err, &apiErr) && apiErr != nil {
		switch apiErr.StatusCode {
		case http.StatusBadRequest, http.StatusPreconditionFailed, http.StatusUnprocessableEntity:
			return fault.Validation(nonEmpty(apiErr.Message, "ZITADEL request is invalid"), fault.Code("PLATFORM_INVALID_INPUT"))
		case http.StatusNotFound:
			return fault.NotFound("ZITADEL user not found", fault.Code("PLATFORM_ZITADEL_USER_NOT_FOUND"))
		case http.StatusConflict:
			return fault.Conflict(nonEmpty(apiErr.Message, "ZITADEL request conflicted"), fault.Code("PLATFORM_ZITADEL_CONFLICT"))
		case http.StatusTooManyRequests:
			return fault.TooManyRequests("ZITADEL admin API rate limit exceeded", fault.Code("PLATFORM_ZITADEL_RATE_LIMIT"))
		case http.StatusUnauthorized, http.StatusForbidden:
			return fault.Unavailable("ZITADEL admin token is invalid or missing required permissions", fault.Code("PLATFORM_ZITADEL_UNAVAILABLE"))
		default:
			if apiErr.StatusCode >= 500 {
				return fault.Unavailable("ZITADEL admin API is unavailable", fault.Code("PLATFORM_ZITADEL_UNAVAILABLE"))
			}
		}
	}
	return fault.Internal("Force verify ZITADEL email failed", fault.Code("INTERNAL"), fault.WithCause(err))
}

func containsRole(roles []domain.Role, target domain.Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func canTransitionOrganizationReviews(roles []domain.Role) bool {
	for _, role := range roles {
		if role.CanTransitionOrganizationReviews() {
			return true
		}
	}
	return false
}

func summarizeRoles(label string, roles []domain.Role) string {
	values := domain.RoleStrings(roles)
	if len(values) == 0 {
		return label + "=[]"
	}
	return label + "=[" + strings.Join(values, ",") + "]"
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func derefUUID(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func createdID(rule *domain.AutoGrantRule) *uuid.UUID {
	if rule == nil {
		return nil
	}
	return rule.ID
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
