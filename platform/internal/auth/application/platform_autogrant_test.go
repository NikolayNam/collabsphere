package application

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/google/uuid"
)

type fakePlatformRoleGrantRepo struct {
	calls         int
	accountID     uuid.UUID
	grantedByID   *uuid.UUID
	grantedAt     time.Time
	roles         []string
	matchedRoles  []string
	matchCalls    int
	matchSubject  string
	matchEmail    string
	matchVerified bool
	returnErr     error
}

func (f *fakePlatformRoleGrantRepo) EnsurePlatformRoles(ctx context.Context, accountID uuid.UUID, roles []string, grantedByAccountID *uuid.UUID, now time.Time) error {
	f.calls++
	f.accountID = accountID
	f.grantedByID = grantedByAccountID
	f.grantedAt = now
	f.roles = append([]string{}, roles...)
	return f.returnErr
}

func (f *fakePlatformRoleGrantRepo) MatchPlatformRoles(ctx context.Context, subject string, email string, emailVerified bool) ([]string, error) {
	f.matchCalls++
	f.matchSubject = subject
	f.matchEmail = email
	f.matchVerified = emailVerified
	return append([]string{}, f.matchedRoles...), nil
}

func TestOIDCPlatformAutoGrantPolicyMatchesNormalizedEmail(t *testing.T) {
	policy := newOIDCPlatformAutoGrantPolicy(OIDCPlatformAutoGrantPolicy{
		PlatformAdminEmails: []string{"ADMIN@collabsphere.ru"},
	})
	identity := &ports.OIDCIdentity{Email: "admin@collabsphere.ru", EmailVerified: true}
	got := policy.ResolveRoles(identity)
	want := []string{"platform_admin"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveRoles() = %v, want %v", got, want)
	}
}

func TestOIDCPlatformAutoGrantPolicyMatchesMultipleRoles(t *testing.T) {
	policy := newOIDCPlatformAutoGrantPolicy(OIDCPlatformAutoGrantPolicy{
		SupportOperatorEmails:  []string{"support@collabsphere.ru"},
		ReviewOperatorSubjects: []string{"zitadel-reviewer-1"},
	})
	identity := &ports.OIDCIdentity{Subject: "zitadel-reviewer-1", Email: "support@collabsphere.ru", EmailVerified: true}
	got := policy.ResolveRoles(identity)
	want := []string{"review_operator", "support_operator"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveRoles() = %v, want %v", got, want)
	}
}

func TestOIDCPlatformAutoGrantPolicyRejectsUnverifiedEmail(t *testing.T) {
	policy := newOIDCPlatformAutoGrantPolicy(OIDCPlatformAutoGrantPolicy{
		PlatformAdminEmails: []string{"admin@collabsphere.ru"},
	})
	identity := &ports.OIDCIdentity{Email: "admin@collabsphere.ru", EmailVerified: false}
	if got := policy.ResolveRoles(identity); len(got) != 0 {
		t.Fatalf("ResolveRoles() = %v, want no roles for unverified email", got)
	}
}

func TestAutoGrantPlatformRolesMergesStaticAndDatabaseRules(t *testing.T) {
	repo := &fakePlatformRoleGrantRepo{matchedRoles: []string{"review_operator"}}
	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	accountID := uuid.MustParse("ee158ed2-740c-430b-bc7a-8fcb0d7eacbf")
	flow := &oidcFlow{
		platformRoles: repo,
		autoGrantPolicy: newOIDCPlatformAutoGrantPolicy(OIDCPlatformAutoGrantPolicy{
			PlatformAdminEmails: []string{"admin@collabsphere.ru"},
		}),
	}

	err := flow.autoGrantPlatformRoles(context.Background(), accountID, &ports.OIDCIdentity{
		Subject:       "zitadel-user-1",
		Email:         "admin@collabsphere.ru",
		EmailVerified: true,
	}, now)
	if err != nil {
		t.Fatalf("autoGrantPlatformRoles() error = %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("EnsurePlatformRoles() calls = %d, want 1", repo.calls)
	}
	wantRoles := []string{"platform_admin", "review_operator"}
	if !reflect.DeepEqual(repo.roles, wantRoles) {
		t.Fatalf("EnsurePlatformRoles() roles = %v, want %v", repo.roles, wantRoles)
	}
	if repo.accountID != accountID {
		t.Fatalf("EnsurePlatformRoles() accountID = %s, want %s", repo.accountID, accountID)
	}
}
