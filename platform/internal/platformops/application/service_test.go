package application

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type fakeTxManager struct{}

func (fakeTxManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeRoleRepo struct {
	rolesByAccount map[uuid.UUID][]domain.Role
	adminIDs       []uuid.UUID
	replaceCalls   int
	lastReplace    []domain.Role
}

func (f *fakeRoleRepo) ListRoles(ctx context.Context, accountID uuid.UUID) ([]domain.Role, error) {
	return append([]domain.Role{}, f.rolesByAccount[accountID]...), nil
}
func (f *fakeRoleRepo) ListAccountIDsByRole(ctx context.Context, role domain.Role) ([]uuid.UUID, error) {
	if role != domain.RolePlatformAdmin {
		return nil, nil
	}
	return append([]uuid.UUID{}, f.adminIDs...), nil
}
func (f *fakeRoleRepo) ReplaceRoles(ctx context.Context, accountID uuid.UUID, roles []domain.Role, grantedByAccountID *uuid.UUID, now time.Time) error {
	f.replaceCalls++
	f.lastReplace = append([]domain.Role{}, roles...)
	f.rolesByAccount[accountID] = append([]domain.Role{}, roles...)
	return nil
}

type fakeAutoGrantRepo struct {
	rules       []domain.AutoGrantRule
	createCalls int
	deleteCalls int
}

func (f *fakeAutoGrantRepo) ListAutoGrantRules(ctx context.Context) ([]domain.AutoGrantRule, error) {
	return append([]domain.AutoGrantRule{}, f.rules...), nil
}

func (f *fakeAutoGrantRepo) CreateAutoGrantRule(ctx context.Context, role domain.Role, matchType domain.AutoGrantMatchType, matchValue string, grantedByAccountID *uuid.UUID, now time.Time) (*domain.AutoGrantRule, error) {
	f.createCalls++
	id := uuid.New()
	createdAt := now
	updatedAt := now
	rule := domain.AutoGrantRule{
		ID:                 &id,
		Role:               role,
		MatchType:          matchType,
		MatchValue:         matchValue,
		Source:             domain.AutoGrantSourceDatabase,
		CreatedByAccountID: grantedByAccountID,
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}
	f.rules = append(f.rules, rule)
	return &rule, nil
}

func (f *fakeAutoGrantRepo) DeleteAutoGrantRule(ctx context.Context, ruleID uuid.UUID) (*domain.AutoGrantRule, error) {
	f.deleteCalls++
	for idx, rule := range f.rules {
		if rule.ID != nil && *rule.ID == ruleID {
			deleted := rule
			f.rules = append(f.rules[:idx], f.rules[idx+1:]...)
			return &deleted, nil
		}
	}
	return nil, nil
}

type fakeAuditRepo struct {
	events []domain.AuditEvent
}

func (f *fakeAuditRepo) Append(ctx context.Context, event domain.AuditEvent) error {
	f.events = append(f.events, event)
	return nil
}

type fakeAccountReader struct {
	accounts map[uuid.UUID]*accdomain.Account
}

func (f *fakeAccountReader) GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error) {
	return f.accounts[id.UUID()], nil
}

type fakeDashboardReader struct{}

func (fakeDashboardReader) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	return &domain.DashboardSummary{}, nil
}

type fakeUploadReader struct{}

func (fakeUploadReader) ListUploadQueue(ctx context.Context, query domain.UploadQueueQuery) ([]domain.UploadQueueItem, int, error) {
	return nil, 0, nil
}

type fakeReviewRepo struct {
	queue                  []domain.OrganizationReviewQueueItem
	total                  int
	detail                 *domain.OrganizationReviewDetail
	updated                *domain.OrganizationReviewCooperationApplication
	updatedLegalDocument   *domain.OrganizationReviewLegalDocument
	lastQueueQuery         domain.OrganizationReviewQueueQuery
	lastPatch              domain.CooperationApplicationReviewPatch
	lastLegalDocumentPatch domain.LegalDocumentReviewPatch
	lastOrganization       uuid.UUID
	lastDocumentID         uuid.UUID
}

func (f *fakeReviewRepo) ListOrganizationReviewQueue(ctx context.Context, query domain.OrganizationReviewQueueQuery) ([]domain.OrganizationReviewQueueItem, int, error) {
	f.lastQueueQuery = query
	return append([]domain.OrganizationReviewQueueItem{}, f.queue...), f.total, nil
}

func (f *fakeReviewRepo) GetOrganizationReview(ctx context.Context, organizationID uuid.UUID) (*domain.OrganizationReviewDetail, error) {
	f.lastOrganization = organizationID
	return f.detail, nil
}

func (f *fakeReviewRepo) UpdateCooperationApplicationReview(ctx context.Context, organizationID uuid.UUID, patch domain.CooperationApplicationReviewPatch) (*domain.OrganizationReviewCooperationApplication, error) {
	f.lastOrganization = organizationID
	f.lastPatch = patch
	return f.updated, nil
}

func (f *fakeReviewRepo) UpdateLegalDocumentReview(ctx context.Context, organizationID, documentID uuid.UUID, patch domain.LegalDocumentReviewPatch) (*domain.OrganizationReviewLegalDocument, error) {
	f.lastOrganization = organizationID
	f.lastDocumentID = documentID
	f.lastLegalDocumentPatch = patch
	return f.updatedLegalDocument, nil
}

type fakeZitadelAdminClient struct {
	result *authports.ZitadelUserEmailVerificationResult
	err    error
}

func (f fakeZitadelAdminClient) ForceVerifyUserEmail(ctx context.Context, userID string) (*authports.ZitadelUserEmailVerificationResult, error) {
	return f.result, f.err
}

type nilAwareZitadelAdminClient struct{}

func (c *nilAwareZitadelAdminClient) ForceVerifyUserEmail(ctx context.Context, userID string) (*authports.ZitadelUserEmailVerificationResult, error) {
	if c == nil {
		return nil, stderrors.New("zitadel admin client is nil")
	}
	return nil, nil
}

func newTestService(roleRepo *fakeRoleRepo, autoGrantRepo *fakeAutoGrantRepo, auditRepo *fakeAuditRepo, accountReader *fakeAccountReader, reviewRepo *fakeReviewRepo, zitadel authports.ZitadelAdminClient, bootstrap []uuid.UUID) *Service {
	return New(roleRepo, autoGrantRepo, auditRepo, accountReader, fakeDashboardReader{}, fakeUploadReader{}, reviewRepo, fakeClock{now: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)}, fakeTxManager{}, zitadel, bootstrap)
}

func mustAccount(t *testing.T, id uuid.UUID, email string) *accdomain.Account {
	t.Helper()
	addr, err := accdomain.NewEmail(email)
	if err != nil {
		t.Fatalf("ParseEmail() error = %v", err)
	}
	accountID, err := accdomain.AccountIDFromUUID(id)
	if err != nil {
		t.Fatalf("AccountIDFromUUID() error = %v", err)
	}
	acc, err := accdomain.RehydrateAccount(accdomain.RehydrateAccountParams{
		ID:        accountID,
		Email:     addr,
		IsActive:  true,
		CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RehydrateAccount() error = %v", err)
	}
	return acc
}

func TestResolveAccessIncludesBootstrapAdmin(t *testing.T) {
	accountID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{accountID: {domain.RoleSupportOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{accountID: mustAccount(t, accountID, "bootstrap@example.com")}}

	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil, []uuid.UUID{accountID})

	access, err := svc.ResolveAccess(context.Background(), accountID)
	if err != nil {
		t.Fatalf("ResolveAccess() error = %v", err)
	}
	if !access.BootstrapAdmin {
		t.Fatalf("BootstrapAdmin = false, want true")
	}
	got := domain.RoleStrings(access.EffectiveRoles)
	want := []string{"platform_admin", "support_operator"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("EffectiveRoles = %v, want %v", got, want)
	}
}

func TestReplaceAccountRolesRejectsRemovingLastAdmin(t *testing.T) {
	actorID := uuid.New()
	targetID := uuid.New()
	roleRepo := &fakeRoleRepo{
		rolesByAccount: map[uuid.UUID][]domain.Role{targetID: {domain.RolePlatformAdmin}},
		adminIDs:       []uuid.UUID{targetID},
	}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{
		actorID:  mustAccount(t, actorID, "actor@example.com"),
		targetID: mustAccount(t, targetID, "target@example.com"),
	}}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil, nil)

	_, err := svc.ReplaceAccountRoles(context.Background(), ReplaceAccountRolesCmd{
		ActorAccountID:  actorID,
		ActorRoles:      []domain.Role{domain.RolePlatformAdmin},
		TargetAccountID: targetID,
		Roles:           nil,
	})
	if err == nil {
		t.Fatal("ReplaceAccountRoles() error = nil, want conflict")
	}
	if roleRepo.replaceCalls != 0 {
		t.Fatalf("ReplaceRoles() calls = %d, want 0", roleRepo.replaceCalls)
	}
}

func TestReplaceAccountRolesWritesAuditAndStoredRoles(t *testing.T) {
	actorID := uuid.New()
	targetID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{
		actorID:  mustAccount(t, actorID, "actor@example.com"),
		targetID: mustAccount(t, targetID, "target@example.com"),
	}}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil, nil)

	access, err := svc.ReplaceAccountRoles(context.Background(), ReplaceAccountRolesCmd{
		ActorAccountID:  actorID,
		ActorRoles:      []domain.Role{domain.RolePlatformAdmin},
		TargetAccountID: targetID,
		Roles:           []string{"support_operator", "review_operator"},
	})
	if err != nil {
		t.Fatalf("ReplaceAccountRoles() error = %v", err)
	}
	got := domain.RoleStrings(access.StoredRoles)
	want := []string{"support_operator", "review_operator"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("StoredRoles = %v, want %v", got, want)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("audit events = %d, want 1", len(auditRepo.events))
	}
	if auditRepo.events[0].Status != domain.AuditStatusSuccess {
		t.Fatalf("audit status = %q, want %q", auditRepo.events[0].Status, domain.AuditStatusSuccess)
	}
}

func TestAddAutoGrantRuleRejectsBootstrapDuplicate(t *testing.T) {
	actorID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{rules: []domain.AutoGrantRule{{
		Role:       domain.RolePlatformAdmin,
		MatchType:  domain.AutoGrantMatchEmail,
		MatchValue: "admin@collabsphere.ru",
		Source:     domain.AutoGrantSourceBootstrap,
	}}}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "actor@example.com")}}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil, nil)

	_, err := svc.AddAutoGrantRule(context.Background(), AddAutoGrantRuleCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		Role:           "platform_admin",
		MatchType:      "email",
		MatchValue:     "admin@collabsphere.ru",
	})
	if err == nil {
		t.Fatal("AddAutoGrantRule() error = nil, want conflict")
	}
	if autoGrantRepo.createCalls != 0 {
		t.Fatalf("CreateAutoGrantRule() calls = %d, want 0", autoGrantRepo.createCalls)
	}
}

func TestAddAutoGrantRuleCreatesDatabaseRule(t *testing.T) {
	actorID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "actor@example.com")}}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil, nil)

	rule, err := svc.AddAutoGrantRule(context.Background(), AddAutoGrantRuleCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		Role:           "support_operator",
		MatchType:      "email",
		MatchValue:     "Support@collabsphere.ru",
	})
	if err != nil {
		t.Fatalf("AddAutoGrantRule() error = %v", err)
	}
	if rule.Source != domain.AutoGrantSourceDatabase || rule.MatchValue != "support@collabsphere.ru" {
		t.Fatalf("AddAutoGrantRule() rule = %+v", rule)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Action != "platform.access.auto_grant_rule.create" {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestForceVerifyUserEmailWithTypedNilClientReturnsDisabled(t *testing.T) {
	actorID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "actor@example.com")}}
	var client *nilAwareZitadelAdminClient
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, client, nil)

	_, err := svc.ForceVerifyUserEmail(context.Background(), ForceVerifyUserEmailCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		UserID:         "123",
	})
	if err == nil {
		t.Fatal("ForceVerifyUserEmail() error = nil, want disabled error")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil {
		t.Fatalf("ForceVerifyUserEmail() error = %v, want fault error", err)
	}
	if appErr.Kind != fault.KindForbidden || appErr.Code != "PLATFORM_ZITADEL_ADMIN_DISABLED" {
		t.Fatalf("ForceVerifyUserEmail() error = kind=%s code=%s, want forbidden/PLATFORM_ZITADEL_ADMIN_DISABLED", appErr.Kind, appErr.Code)
	}
}

func TestForceVerifyUserEmailWritesAudit(t *testing.T) {
	actorID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "actor@example.com")}}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, fakeZitadelAdminClient{result: &authports.ZitadelUserEmailVerificationResult{UserID: "123", Email: "user@example.com", AlreadyVerified: false}}, nil)

	res, err := svc.ForceVerifyUserEmail(context.Background(), ForceVerifyUserEmailCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		UserID:         "123",
	})
	if err != nil {
		t.Fatalf("ForceVerifyUserEmail() error = %v", err)
	}
	if !res.Verified || res.Email != "user@example.com" {
		t.Fatalf("ForceVerifyUserEmail() result = %+v", res)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("audit events = %d, want 1", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "platform.user.email.force_verify" {
		t.Fatalf("audit action = %q, want force-verify", auditRepo.events[0].Action)
	}
}

func TestTransitionCooperationApplicationReviewSubmittedToUnderReview(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleReviewOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "reviewer@example.com")}}
	reviewRepo := &fakeReviewRepo{
		detail: &domain.OrganizationReviewDetail{
			CooperationApplication: &domain.OrganizationReviewCooperationApplication{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				Status:         string(orgdomain.CooperationApplicationStatusSubmitted),
			},
		},
		updated: &domain.OrganizationReviewCooperationApplication{
			ID:                uuid.New(),
			OrganizationID:    organizationID,
			Status:            string(orgdomain.CooperationApplicationStatusUnderReview),
			ReviewerAccountID: &actorID,
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	result, err := svc.TransitionCooperationApplicationReview(context.Background(), TransitionCooperationApplicationReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		OrganizationID: organizationID,
		TargetStatus:   "under_review",
	})
	if err != nil {
		t.Fatalf("TransitionCooperationApplicationReview() error = %v", err)
	}
	if result == nil || result.Status != string(orgdomain.CooperationApplicationStatusUnderReview) {
		t.Fatalf("TransitionCooperationApplicationReview() result = %+v", result)
	}
	if reviewRepo.lastPatch.Status != string(orgdomain.CooperationApplicationStatusUnderReview) {
		t.Fatalf("last patch status = %q", reviewRepo.lastPatch.Status)
	}
	if reviewRepo.lastPatch.ReviewerAccountID == nil || *reviewRepo.lastPatch.ReviewerAccountID != actorID {
		t.Fatalf("last patch reviewer = %v, want %s", reviewRepo.lastPatch.ReviewerAccountID, actorID)
	}
	if reviewRepo.lastPatch.ReviewedAt != nil {
		t.Fatalf("ReviewedAt = %v, want nil for under_review", reviewRepo.lastPatch.ReviewedAt)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Action != "platform.organization.review.transition" || auditRepo.events[0].Status != domain.AuditStatusSuccess {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionCooperationApplicationReviewRequiresNoteForRejected(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleReviewOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "reviewer@example.com")}}
	reviewRepo := &fakeReviewRepo{}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionCooperationApplicationReview(context.Background(), TransitionCooperationApplicationReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		OrganizationID: organizationID,
		TargetStatus:   "rejected",
	})
	if err == nil {
		t.Fatal("TransitionCooperationApplicationReview() error = nil, want validation")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindValidation {
		t.Fatalf("TransitionCooperationApplicationReview() error = %v, want validation fault", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusFailed {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionCooperationApplicationReviewRejectsSupportOperator(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleSupportOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "support@example.com")}}
	reviewRepo := &fakeReviewRepo{}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionCooperationApplicationReview(context.Background(), TransitionCooperationApplicationReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleSupportOperator},
		OrganizationID: organizationID,
		TargetStatus:   "under_review",
	})
	if err == nil {
		t.Fatal("TransitionCooperationApplicationReview() error = nil, want forbidden")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindForbidden {
		t.Fatalf("TransitionCooperationApplicationReview() error = %v, want forbidden fault", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusDenied {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionCooperationApplicationReviewRejectsInvalidTransition(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "admin@example.com")}}
	reviewRepo := &fakeReviewRepo{
		detail: &domain.OrganizationReviewDetail{
			CooperationApplication: &domain.OrganizationReviewCooperationApplication{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				Status:         string(orgdomain.CooperationApplicationStatusDraft),
			},
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionCooperationApplicationReview(context.Background(), TransitionCooperationApplicationReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		OrganizationID: organizationID,
		TargetStatus:   "approved",
	})
	if err == nil {
		t.Fatal("TransitionCooperationApplicationReview() error = nil, want conflict")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindConflict || appErr.Code != "PLATFORM_REVIEW_TRANSITION_INVALID" {
		t.Fatalf("TransitionCooperationApplicationReview() error = %v, want conflict/PLATFORM_REVIEW_TRANSITION_INVALID", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusFailed {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestGetOrganizationReviewEnrichesLegalDocumentVerification(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	documentID := uuid.New()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleReviewOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "reviewer@example.com")}}
	reviewRepo := &fakeReviewRepo{
		detail: &domain.OrganizationReviewDetail{
			Organization: domain.OrganizationReviewOrganization{
				ID: organizationID,
			},
			CooperationApplication: &domain.OrganizationReviewCooperationApplication{
				ID:                    uuid.New(),
				OrganizationID:        organizationID,
				Status:                "submitted",
				ConfirmationEmail:     strPtr("confirm@example.com"),
				CompanyName:           strPtr("Acme"),
				RepresentedCategories: strPtr("Food"),
				MinimumOrderAmount:    strPtr("1000"),
				DeliveryGeography:     strPtr("Moscow"),
				SalesChannels:         []string{"Retail"},
				PriceListObjectID:     uuidPtr(uuid.New()),
				ContactFirstName:      strPtr("Ivan"),
				ContactLastName:       strPtr("Petrov"),
				ContactJobTitle:       strPtr("Manager"),
				ContactEmail:          strPtr("sales@example.com"),
				ContactPhone:          strPtr("+79990000000"),
			},
			LegalDocuments: []domain.OrganizationReviewLegalDocument{{
				ID:             documentID,
				OrganizationID: organizationID,
				DocumentType:   "charter",
				Status:         "pending",
				Analysis: &domain.OrganizationReviewLegalDocumentAnalysis{
					ID:                   uuid.New(),
					DocumentID:           documentID,
					OrganizationID:       organizationID,
					Status:               "completed",
					ExtractedFieldsJSON:  json.RawMessage(`{"companyName":"Acme"}`),
					DetectedDocumentType: strPtr("charter"),
					ConfidenceScore:      floatPtr(0.99),
					RequestedAt:          now,
					UpdatedAt:            timePtr(now),
				},
			}},
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	detail, err := svc.GetOrganizationReview(context.Background(), organizationID)
	if err != nil {
		t.Fatalf("GetOrganizationReview() error = %v", err)
	}
	if detail == nil || len(detail.LegalDocuments) != 1 {
		t.Fatalf("GetOrganizationReview() detail = %+v", detail)
	}
	verification := detail.LegalDocuments[0].Verification
	if verification == nil {
		t.Fatal("verification = nil, want shared verifier output")
	}
	if verification.Verdict != "approved" {
		t.Fatalf("verification verdict = %q, want approved", verification.Verdict)
	}
	if len(verification.MissingFields) != 0 {
		t.Fatalf("verification missing fields = %v, want empty", verification.MissingFields)
	}
	if detail.KYC == nil {
		t.Fatal("KYC = nil, want aggregated snapshot")
	}
	if detail.KYC.Status != "pending_verification" {
		t.Fatalf("KYC status = %q, want pending_verification", detail.KYC.Status)
	}
}

func TestTransitionLegalDocumentReviewPendingToApproved(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	documentID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleReviewOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "reviewer@example.com")}}
	reviewRepo := &fakeReviewRepo{
		detail: &domain.OrganizationReviewDetail{
			LegalDocuments: []domain.OrganizationReviewLegalDocument{{
				ID:             documentID,
				OrganizationID: organizationID,
				Status:         string(orgdomain.OrganizationLegalDocumentStatusPending),
			}},
		},
		updatedLegalDocument: &domain.OrganizationReviewLegalDocument{
			ID:                documentID,
			OrganizationID:    organizationID,
			Status:            string(orgdomain.OrganizationLegalDocumentStatusApproved),
			ReviewerAccountID: &actorID,
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	result, err := svc.TransitionLegalDocumentReview(context.Background(), TransitionLegalDocumentReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		OrganizationID: organizationID,
		DocumentID:     documentID,
		TargetStatus:   "approved",
	})
	if err != nil {
		t.Fatalf("TransitionLegalDocumentReview() error = %v", err)
	}
	if result == nil || result.Status != string(orgdomain.OrganizationLegalDocumentStatusApproved) {
		t.Fatalf("TransitionLegalDocumentReview() result = %+v", result)
	}
	if reviewRepo.lastDocumentID != documentID {
		t.Fatalf("last document id = %s, want %s", reviewRepo.lastDocumentID, documentID)
	}
	if reviewRepo.lastLegalDocumentPatch.Status != string(orgdomain.OrganizationLegalDocumentStatusApproved) {
		t.Fatalf("last legal patch status = %q", reviewRepo.lastLegalDocumentPatch.Status)
	}
	if reviewRepo.lastLegalDocumentPatch.ReviewerAccountID == nil || *reviewRepo.lastLegalDocumentPatch.ReviewerAccountID != actorID {
		t.Fatalf("last legal patch reviewer = %v, want %s", reviewRepo.lastLegalDocumentPatch.ReviewerAccountID, actorID)
	}
	if reviewRepo.lastLegalDocumentPatch.ReviewedAt == nil || reviewRepo.lastLegalDocumentPatch.ReviewedAt.IsZero() {
		t.Fatalf("ReviewedAt = %v, want timestamp", reviewRepo.lastLegalDocumentPatch.ReviewedAt)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Action != "platform.organization.legal_document.review.transition" || auditRepo.events[0].Status != domain.AuditStatusSuccess {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionLegalDocumentReviewRequiresNoteForRejected(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	documentID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleReviewOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "reviewer@example.com")}}
	reviewRepo := &fakeReviewRepo{}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionLegalDocumentReview(context.Background(), TransitionLegalDocumentReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleReviewOperator},
		OrganizationID: organizationID,
		DocumentID:     documentID,
		TargetStatus:   "rejected",
	})
	if err == nil {
		t.Fatal("TransitionLegalDocumentReview() error = nil, want validation")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindValidation {
		t.Fatalf("TransitionLegalDocumentReview() error = %v, want validation fault", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusFailed {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionLegalDocumentReviewRejectsSupportOperator(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	documentID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RoleSupportOperator}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "support@example.com")}}
	reviewRepo := &fakeReviewRepo{}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionLegalDocumentReview(context.Background(), TransitionLegalDocumentReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RoleSupportOperator},
		OrganizationID: organizationID,
		DocumentID:     documentID,
		TargetStatus:   "approved",
	})
	if err == nil {
		t.Fatal("TransitionLegalDocumentReview() error = nil, want forbidden")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindForbidden {
		t.Fatalf("TransitionLegalDocumentReview() error = %v, want forbidden fault", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusDenied {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func TestTransitionLegalDocumentReviewRejectsNoOpTransition(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	documentID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}}
	auditRepo := &fakeAuditRepo{}
	autoGrantRepo := &fakeAutoGrantRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "admin@example.com")}}
	reviewRepo := &fakeReviewRepo{
		detail: &domain.OrganizationReviewDetail{
			LegalDocuments: []domain.OrganizationReviewLegalDocument{{
				ID:             documentID,
				OrganizationID: organizationID,
				Status:         string(orgdomain.OrganizationLegalDocumentStatusApproved),
			}},
		},
	}
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, reviewRepo, nil, nil)

	_, err := svc.TransitionLegalDocumentReview(context.Background(), TransitionLegalDocumentReviewCmd{
		ActorAccountID: actorID,
		ActorRoles:     []domain.Role{domain.RolePlatformAdmin},
		OrganizationID: organizationID,
		DocumentID:     documentID,
		TargetStatus:   "approved",
	})
	if err == nil {
		t.Fatal("TransitionLegalDocumentReview() error = nil, want conflict")
	}
	appErr, ok := fault.As(err)
	if !ok || appErr == nil || appErr.Kind != fault.KindConflict || appErr.Code != "PLATFORM_REVIEW_TRANSITION_INVALID" {
		t.Fatalf("TransitionLegalDocumentReview() error = %v, want conflict/PLATFORM_REVIEW_TRANSITION_INVALID", err)
	}
	if len(auditRepo.events) != 1 || auditRepo.events[0].Status != domain.AuditStatusFailed {
		t.Fatalf("audit events = %+v", auditRepo.events)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func strPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	return &value
}
