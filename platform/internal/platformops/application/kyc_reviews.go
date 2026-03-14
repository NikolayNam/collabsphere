package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	sharedkyc "github.com/NikolayNam/collabsphere/shared/kyc"
	"github.com/google/uuid"
)

type ListKYCReviewsCmd struct {
	Scope  *string
	Status *string
	Limit  int
	Offset int
}

type GetKYCReviewCmd struct {
	ReviewID string
}

type DecideKYCReviewCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	ReviewID       string
	Decision       string
	Reason         *string
}

type DecideKYCDocumentReviewCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	ReviewID       string
	DocumentID     uuid.UUID
	Decision       string
	Reason         *string
}

type ListKYCLevelsCmd struct {
	Scope *string
}

type UpsertKYCLevelCmd struct {
	ID                    *uuid.UUID
	Scope                 string
	Code                  string
	Name                  string
	Rank                  int
	IsActive              bool
	RequiredDocumentTypes []domain.KYCLevelRequirement
}

type DeleteKYCLevelCmd struct {
	LevelID uuid.UUID
}

type IssueKYCLevelCmd struct {
	ActorAccountID uuid.UUID
	ActorRoles     []domain.Role
	ActorBootstrap bool
	ReviewID       string
}

func (s *Service) ListKYCReviews(ctx context.Context, cmd ListKYCReviewsCmd) ([]domain.KYCReviewItem, int, error) {
	scope := normalizeScope(cmd.Scope)
	status := normalizeStatus(cmd.Status)
	return s.reviews.ListKYCReviews(ctx, domain.KYCReviewQuery{
		Scope:  scope,
		Status: status,
		Limit:  cmd.Limit,
		Offset: cmd.Offset,
	})
}

func (s *Service) GetKYCReview(ctx context.Context, cmd GetKYCReviewCmd) (*domain.KYCReviewDetail, error) {
	scope, subjectID, err := parseReviewID(cmd.ReviewID)
	if err != nil {
		return nil, err
	}
	detail, err := s.reviews.GetKYCReview(ctx, scope, subjectID)
	if err != nil {
		return nil, fault.Internal("Load KYC review failed", fault.WithCause(err))
	}
	if detail == nil {
		return nil, fault.NotFound("KYC review not found", fault.Code("PLATFORM_KYC_REVIEW_NOT_FOUND"))
	}
	documents, err := s.reviews.ListKYCDocuments(ctx, scope, subjectID)
	if err != nil {
		return nil, fault.Internal("Load KYC documents failed", fault.WithCause(err))
	}
	events, err := s.reviews.ListKYCReviewEvents(ctx, scope, subjectID, 50)
	if err != nil {
		return nil, fault.Internal("Load KYC review events failed", fault.WithCause(err))
	}
	detail.Documents = documents
	detail.Events = events
	return detail, nil
}

func (s *Service) DecideKYCReview(ctx context.Context, cmd DecideKYCReviewCmd) (*domain.KYCReviewDetail, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	scope, subjectID, err := parseReviewID(cmd.ReviewID)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.review.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	decision, ok := sharedkyc.ParseDecision(cmd.Decision)
	if !ok {
		err := fault.Validation("KYC decision is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("decision", "must be approve|reject|request_info"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.review.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	targetStatus := sharedkyc.StatusApproved
	switch decision {
	case sharedkyc.DecisionApprove:
		targetStatus = sharedkyc.StatusApproved
	case sharedkyc.DecisionReject:
		targetStatus = sharedkyc.StatusRejected
	case sharedkyc.DecisionRequestInfo:
		targetStatus = sharedkyc.StatusNeedsInfo
	}
	now := s.clock.Now()
	detail, err := s.reviews.ApplyKYCDecision(ctx, domain.KYCDecisionPatch{
		Scope:             scope,
		SubjectID:         subjectID,
		Status:            string(targetStatus),
		ReviewNote:        cleanReason(cmd.Reason),
		ReviewerAccountID: cmd.ActorAccountID,
		ReviewedAt:        now,
		UpdatedAt:         now,
	})
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.review.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, fault.Internal("Apply KYC decision failed", fault.WithCause(err))
	}
	if detail == nil {
		err := fault.NotFound("KYC review not found", fault.Code("PLATFORM_KYC_REVIEW_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.review.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	if err := s.reviews.AppendKYCReviewEvent(ctx, domain.KYCReviewEvent{
		ID:                uuid.New(),
		Scope:             scope,
		SubjectID:         subjectID,
		Decision:          string(decision),
		Reason:            cleanReason(cmd.Reason),
		ReviewerAccountID: cmd.ActorAccountID,
		CreatedAt:         now,
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.review.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, fault.Internal("Append KYC review event failed", fault.WithCause(err))
	}
	s.appendAuditBestEffort(ctx, domain.AuditEvent{
		ActorAccountID: &cmd.ActorAccountID,
		ActorRoles:     cmd.ActorRoles,
		ActorBootstrap: cmd.ActorBootstrap,
		Action:         "platform.kyc.review.decision",
		TargetType:     "kyc_review",
		TargetID:       stringPtr(cmd.ReviewID),
		Status:         domain.AuditStatusSuccess,
		Summary:        stringPtr(fmt.Sprintf("decision=%s status=%s", decision, targetStatus)),
		CreatedAt:      now,
	})
	return detail, nil
}

func (s *Service) DecideKYCDocumentReview(ctx context.Context, cmd DecideKYCDocumentReviewCmd) (*domain.KYCReviewDetail, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	if cmd.DocumentID == uuid.Nil {
		return nil, fault.Validation("Document id is invalid", fault.Field("documentId", "must be a UUID"))
	}
	scope, subjectID, err := parseReviewID(cmd.ReviewID)
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.document.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	decision, ok := sharedkyc.ParseDecision(cmd.Decision)
	if !ok {
		err := fault.Validation("KYC document decision is invalid", fault.Code("PLATFORM_INVALID_INPUT"), fault.Field("decision", "must be approve|reject|request_info"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.document.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	targetStatus := sharedkyc.DocumentStatusVerified
	switch decision {
	case sharedkyc.DecisionApprove:
		targetStatus = sharedkyc.DocumentStatusVerified
	case sharedkyc.DecisionReject, sharedkyc.DecisionRequestInfo:
		targetStatus = sharedkyc.DocumentStatusRejected
	}
	now := s.clock.Now()
	updated, err := s.reviews.ApplyKYCDocumentDecision(ctx, domain.KYCDocumentDecisionPatch{
		Scope:             scope,
		SubjectID:         subjectID,
		DocumentID:        cmd.DocumentID,
		Status:            string(targetStatus),
		ReviewNote:        cleanReason(cmd.Reason),
		ReviewerAccountID: cmd.ActorAccountID,
		ReviewedAt:        now,
		UpdatedAt:         now,
	})
	if err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.document.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, fault.Internal("Apply KYC document decision failed", fault.WithCause(err))
	}
	if updated == nil {
		err := fault.NotFound("KYC document not found", fault.Code("PLATFORM_KYC_DOCUMENT_NOT_FOUND"))
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.document.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, err
	}
	documentReason := cleanReason(cmd.Reason)
	if documentReason != nil {
		text := fmt.Sprintf("document:%s %s", cmd.DocumentID.String(), *documentReason)
		documentReason = &text
	} else {
		text := fmt.Sprintf("document:%s", cmd.DocumentID.String())
		documentReason = &text
	}
	if err := s.reviews.AppendKYCReviewEvent(ctx, domain.KYCReviewEvent{
		ID:                uuid.New(),
		Scope:             scope,
		SubjectID:         subjectID,
		Decision:          string(decision),
		Reason:            documentReason,
		ReviewerAccountID: cmd.ActorAccountID,
		CreatedAt:         now,
	}); err != nil {
		s.recordFailure(ctx, cmd.ActorAccountID, cmd.ActorRoles, cmd.ActorBootstrap, "platform.kyc.document.decision", "kyc_review", cmd.ReviewID, err.Error())
		return nil, fault.Internal("Append KYC document review event failed", fault.WithCause(err))
	}
	detail, err := s.GetKYCReview(ctx, GetKYCReviewCmd{ReviewID: cmd.ReviewID})
	if err != nil {
		return nil, err
	}
	s.appendAuditBestEffort(ctx, domain.AuditEvent{
		ActorAccountID: &cmd.ActorAccountID,
		ActorRoles:     cmd.ActorRoles,
		ActorBootstrap: cmd.ActorBootstrap,
		Action:         "platform.kyc.document.decision",
		TargetType:     "kyc_review",
		TargetID:       stringPtr(cmd.ReviewID),
		Status:         domain.AuditStatusSuccess,
		Summary:        stringPtr(fmt.Sprintf("document_id=%s decision=%s status=%s", cmd.DocumentID.String(), decision, targetStatus)),
		CreatedAt:      now,
	})
	return detail, nil
}

func (s *Service) ListKYCLevels(ctx context.Context, cmd ListKYCLevelsCmd) ([]domain.KYCLevel, error) {
	scope := normalizeScope(cmd.Scope)
	return s.reviews.ListKYCLevels(ctx, scope)
}

func (s *Service) UpsertKYCLevel(ctx context.Context, cmd UpsertKYCLevelCmd) (*domain.KYCLevel, error) {
	scopeParsed, ok := sharedkyc.ParseScope(cmd.Scope)
	if !ok {
		return nil, fault.Validation("KYC level scope is invalid", fault.Field("scope", "must be account|organization"))
	}
	code := strings.TrimSpace(cmd.Code)
	if code == "" {
		return nil, fault.Validation("KYC level code is required", fault.Field("code", "required"))
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, fault.Validation("KYC level name is required", fault.Field("name", "required"))
	}
	if cmd.Rank <= 0 {
		return nil, fault.Validation("KYC level rank is invalid", fault.Field("rank", "must be > 0"))
	}
	requirements := make([]domain.KYCLevelRequirement, 0, len(cmd.RequiredDocumentTypes))
	for _, item := range cmd.RequiredDocumentTypes {
		documentType := strings.TrimSpace(item.DocumentType)
		if documentType == "" {
			return nil, fault.Validation("KYC level requirement is invalid", fault.Field("requiredDocumentTypes", "documentType required"))
		}
		minCount := item.MinCount
		if minCount <= 0 {
			minCount = 1
		}
		requirements = append(requirements, domain.KYCLevelRequirement{
			DocumentType: documentType,
			MinCount:     minCount,
		})
	}
	now := s.clock.Now()
	id := uuid.Nil
	if cmd.ID != nil {
		id = *cmd.ID
	}
	createdAt := now
	if id == uuid.Nil {
		id = uuid.New()
	}
	scopeValue := string(scopeParsed)
	existing, err := s.reviews.ListKYCLevels(ctx, &scopeValue)
	if err == nil {
		for _, item := range existing {
			if item.ID == id {
				createdAt = item.CreatedAt
				break
			}
		}
	}
	return s.reviews.UpsertKYCLevel(ctx, domain.KYCLevel{
		ID:                    id,
		Scope:                 string(scopeParsed),
		Code:                  code,
		Name:                  name,
		Rank:                  cmd.Rank,
		IsActive:              cmd.IsActive,
		RequiredDocumentTypes: requirements,
		CreatedAt:             createdAt,
		UpdatedAt:             now,
	})
}

func (s *Service) DeleteKYCLevel(ctx context.Context, cmd DeleteKYCLevelCmd) error {
	if cmd.LevelID == uuid.Nil {
		return fault.Validation("KYC level id is invalid", fault.Field("levelId", "must be UUID"))
	}
	return s.reviews.DeleteKYCLevel(ctx, cmd.LevelID)
}

func (s *Service) IssueKYCLevel(ctx context.Context, cmd IssueKYCLevelCmd) (*domain.KYCLevelAssignment, error) {
	if cmd.ActorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required", fault.Code("PLATFORM_UNAUTHORIZED"))
	}
	scope, subjectID, err := parseReviewID(cmd.ReviewID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	assignment, err := s.reviews.EvaluateAndAssignKYCLevel(ctx, scope, subjectID, cmd.ActorAccountID, now)
	if err != nil {
		return nil, fault.Internal("Issue KYC level failed", fault.WithCause(err))
	}
	reason := "issued:none"
	if assignment != nil && assignment.LevelCode != nil {
		reason = "issued:" + *assignment.LevelCode
	}
	_ = s.reviews.AppendKYCReviewEvent(ctx, domain.KYCReviewEvent{
		ID:                uuid.New(),
		Scope:             scope,
		SubjectID:         subjectID,
		Decision:          "approve",
		Reason:            &reason,
		ReviewerAccountID: cmd.ActorAccountID,
		CreatedAt:         now,
	})
	return assignment, nil
}

func parseReviewID(value string) (string, uuid.UUID, error) {
	trimmed := strings.TrimSpace(value)
	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		return "", uuid.Nil, fault.Validation("Review id is invalid", fault.Field("reviewId", "must be '<scope>:<uuid>'"))
	}
	scope, ok := sharedkyc.ParseScope(parts[0])
	if !ok {
		return "", uuid.Nil, fault.Validation("Review id scope is invalid", fault.Field("reviewId", "scope must be account|organization"))
	}
	subjectID, err := uuid.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", uuid.Nil, fault.Validation("Review id subject is invalid", fault.Field("reviewId", "subject must be a UUID"))
	}
	return string(scope), subjectID, nil
}

func normalizeScope(scope *string) *string {
	if scope == nil {
		return nil
	}
	parsed, ok := sharedkyc.ParseScope(*scope)
	if !ok {
		return nil
	}
	value := string(parsed)
	return &value
}

func normalizeStatus(status *string) *string {
	if status == nil {
		return nil
	}
	parsed, ok := sharedkyc.ParseStatus(*status)
	if !ok {
		return nil
	}
	value := string(parsed)
	return &value
}

func cleanReason(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
