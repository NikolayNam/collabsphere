package application

import (
	"context"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
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

func newTestService(roleRepo *fakeRoleRepo, auditRepo *fakeAuditRepo, accountReader *fakeAccountReader, zitadel authports.ZitadelAdminClient, bootstrap []uuid.UUID) *Service {
	return New(roleRepo, auditRepo, accountReader, fakeDashboardReader{}, fakeUploadReader{}, fakeClock{now: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)}, fakeTxManager{}, zitadel, bootstrap)
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
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{accountID: mustAccount(t, accountID, "bootstrap@example.com")}}

	svc := newTestService(roleRepo, auditRepo, accounts, nil, []uuid.UUID{accountID})

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
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{
		actorID:  mustAccount(t, actorID, "actor@example.com"),
		targetID: mustAccount(t, targetID, "target@example.com"),
	}}
	svc := newTestService(roleRepo, auditRepo, accounts, nil, nil)

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
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{
		actorID:  mustAccount(t, actorID, "actor@example.com"),
		targetID: mustAccount(t, targetID, "target@example.com"),
	}}
	svc := newTestService(roleRepo, auditRepo, accounts, nil, nil)

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

func TestForceVerifyUserEmailWritesAudit(t *testing.T) {
	actorID := uuid.New()
	roleRepo := &fakeRoleRepo{rolesByAccount: map[uuid.UUID][]domain.Role{actorID: {domain.RolePlatformAdmin}}, adminIDs: []uuid.UUID{actorID}}
	auditRepo := &fakeAuditRepo{}
	accounts := &fakeAccountReader{accounts: map[uuid.UUID]*accdomain.Account{actorID: mustAccount(t, actorID, "actor@example.com")}}
	svc := newTestService(roleRepo, auditRepo, accounts, fakeZitadelAdminClient{result: &authports.ZitadelUserEmailVerificationResult{UserID: "123", Email: "user@example.com", AlreadyVerified: false}}, nil)

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
