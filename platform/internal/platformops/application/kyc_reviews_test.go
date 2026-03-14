package application

import (
	"context"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
)

func TestDecideKYCReviewApprove(t *testing.T) {
	actorID := uuid.New()
	subjectID := uuid.New()

	roleRepo := &fakeRoleRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	auditRepo := &fakeAuditRepo{}
	accounts := &fakeAccountReader{}
	reviewRepo := &fakeReviewRepo{
		kycDetail: &domain.KYCReviewDetail{
			ReviewID:    "account:" + subjectID.String(),
			Scope:       "account",
			SubjectID:   subjectID,
			Status:      "submitted",
			CreatedAt:   time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
			SubmittedAt: func() *time.Time { v := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC); return &v }(),
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	detail, err := svc.DecideKYCReview(context.Background(), DecideKYCReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		ReviewID:       "account:" + subjectID.String(),
		Decision:       "approve",
	})
	if err != nil {
		t.Fatalf("DecideKYCReview() error = %v", err)
	}
	if detail.Status != "submitted" {
		t.Fatalf("detail status should come from repository, got %s", detail.Status)
	}
	if reviewRepo.lastKYCPatch.Status != "approved" {
		t.Fatalf("applied status = %s, want approved", reviewRepo.lastKYCPatch.Status)
	}
	if reviewRepo.lastKYCPatch.Scope != "account" {
		t.Fatalf("scope = %s, want account", reviewRepo.lastKYCPatch.Scope)
	}
}

func TestGetKYCReviewInvalidID(t *testing.T) {
	svc := newTestService(&fakeRoleRepo{}, &fakeAutoGrantRepo{}, &fakeAuditRepo{}, &fakeAccountReader{}, &fakeReviewRepo{}, nil, nil)
	_, err := svc.GetKYCReview(context.Background(), GetKYCReviewCmd{ReviewID: "bad"})
	if err == nil {
		t.Fatal("expected validation error for invalid review id")
	}
}

func TestDecideKYCDocumentReviewApprove(t *testing.T) {
	actorID := uuid.New()
	subjectID := uuid.New()
	documentID := uuid.New()

	reviewRepo := &fakeReviewRepo{
		kycDetail: &domain.KYCReviewDetail{
			ReviewID:  "organization:" + subjectID.String(),
			Scope:     "organization",
			SubjectID: subjectID,
			Status:    "in_review",
			CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
		},
		kycDocuments: []domain.KYCDocumentReviewItem{{
			ID:        documentID,
			ObjectID:  uuid.New(),
			Status:    "uploaded",
			Title:     "Charter",
			CreatedAt: time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
		}},
	}
	svc := newTestService(&fakeRoleRepo{}, &fakeAutoGrantRepo{}, &fakeAuditRepo{}, &fakeAccountReader{}, reviewRepo, nil, nil)

	detail, err := svc.DecideKYCDocumentReview(context.Background(), DecideKYCDocumentReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		ReviewID:       "organization:" + subjectID.String(),
		DocumentID:     documentID,
		Decision:       "approve",
	})
	if err != nil {
		t.Fatalf("DecideKYCDocumentReview() error = %v", err)
	}
	if detail == nil {
		t.Fatal("DecideKYCDocumentReview() detail = nil")
	}
	if reviewRepo.lastKYCDocumentPatch.Status != "verified" {
		t.Fatalf("document patch status = %s, want verified", reviewRepo.lastKYCDocumentPatch.Status)
	}
	if len(reviewRepo.kycEvents) == 0 {
		t.Fatal("expected review event to be appended")
	}
}
