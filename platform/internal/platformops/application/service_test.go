package application

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
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

func newTestService(roleRepo *fakeRoleRepo, autoGrantRepo *fakeAutoGrantRepo, auditRepo *fakeAuditRepo, accountReader *fakeAccountReader, zitadel authports.ZitadelAdminClient, bootstrap []uuid.UUID) *Service {
	return New(roleRepo, autoGrantRepo, auditRepo, accountReader, fakeDashboardReader{}, fakeUploadReader{}, fakeClock{now: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)}, fakeTxManager{}, zitadel, bootstrap)
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

	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, []uuid.UUID{accountID})

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil)

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil)

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil)

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, nil, nil)

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, client, nil)

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
	svc := newTestService(roleRepo, autoGrantRepo, auditRepo, accounts, fakeZitadelAdminClient{result: &authports.ZitadelUserEmailVerificationResult{UserID: "123", Email: "user@example.com", AlreadyVerified: false}}, nil)

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

