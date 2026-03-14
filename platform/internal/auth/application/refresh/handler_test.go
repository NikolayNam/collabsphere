package refresh

import (
	"context"
	"sync"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

type fakeAccountReader struct {
	account *accdomain.Account
}

func (f *fakeAccountReader) GetByEmail(ctx context.Context, email accdomain.Email) (*accdomain.Account, error) {
	return nil, nil
}

func (f *fakeAccountReader) GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error) {
	if f.account == nil || f.account.ID() != id {
		return nil, nil
	}
	return f.account, nil
}

func (f *fakeAccountReader) Create(ctx context.Context, account *accdomain.Account) error {
	f.account = account
	return nil
}

type fakeTokenManager struct{}

func (fakeTokenManager) GenerateAccessToken(ctx context.Context, principal authdomain.Principal, expiresAt time.Time) (string, error) {
	return "access:" + principal.SessionID.String(), nil
}

func (fakeTokenManager) VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error) {
	return authdomain.Principal{}, nil
}

func (fakeTokenManager) SessionTTL() time.Duration { return 24 * time.Hour }
func (fakeTokenManager) AccessTTL() time.Duration  { return 15 * time.Minute }

type fakeRandom struct {
	mu     sync.Mutex
	values []string
	index  int
}

func (f *fakeRandom) Generate() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	value := f.values[f.index]
	f.index++
	return value, nil
}

func (f *fakeRandom) Hash(raw string) string {
	return "hash:" + raw
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time { return f.now }

type fakeSessionRepo struct {
	mu      sync.Mutex
	session *authdomain.RefreshSession
	known   map[string]uuid.UUID
	used    map[string]bool
}

func newFakeSessionRepo(session *authdomain.RefreshSession) *fakeSessionRepo {
	return &fakeSessionRepo{
		session: session,
		known: map[string]uuid.UUID{
			session.TokenHash(): session.ID(),
		},
		used: make(map[string]bool),
	}
}

func (f *fakeSessionRepo) Create(ctx context.Context, session *authdomain.RefreshSession) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.session = session
	if f.known == nil {
		f.known = make(map[string]uuid.UUID)
	}
	if f.used == nil {
		f.used = make(map[string]bool)
	}
	f.known[session.TokenHash()] = session.ID()
	return nil
}

func (f *fakeSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*authdomain.RefreshSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.session == nil {
		return nil, nil
	}
	if _, ok := f.known[tokenHash]; !ok {
		return nil, nil
	}
	return cloneSessionForTest(takeSessionSnapshot(f.session))
}

func (f *fakeSessionRepo) RotateByRefreshToken(ctx context.Context, presentedTokenHash, newTokenHash string, now time.Time) (*authdomain.RefreshSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.session == nil {
		return nil, nil
	}
	if _, ok := f.known[presentedTokenHash]; !ok {
		return nil, nil
	}
	if f.session.IsRevoked() || f.session.IsExpired(now) {
		return nil, nil
	}
	if f.used[presentedTokenHash] || f.session.TokenHash() != presentedTokenHash {
		_ = f.session.Revoke(now)
		return nil, nil
	}

	f.used[presentedTokenHash] = true
	f.known[newTokenHash] = f.session.ID()

	rotated, err := authdomain.RehydrateRefreshSession(authdomain.RehydrateRefreshSessionParams{
		ID:        f.session.ID(),
		AccountID: f.session.AccountID(),
		TokenHash: newTokenHash,
		UserAgent: f.session.UserAgent(),
		IP:        f.session.IP(),
		ExpiresAt: f.session.ExpiresAt(),
		RevokedAt: f.session.RevokedAt(),
		CreatedAt: f.session.CreatedAt(),
		UpdatedAt: &now,
	})
	if err != nil {
		return nil, err
	}
	f.session = rotated
	return cloneSessionForTest(takeSessionSnapshot(f.session))
}

func (f *fakeSessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.session != nil && f.session.ID() == id {
		_ = f.session.Revoke(time.Now())
	}
	return nil
}

type sessionSnapshot struct {
	id        uuid.UUID
	accountID uuid.UUID
	tokenHash string
	userAgent *string
	ip        *string
	expiresAt time.Time
	revokedAt *time.Time
	createdAt time.Time
	updatedAt *time.Time
}

func takeSessionSnapshot(session *authdomain.RefreshSession) sessionSnapshot {
	return sessionSnapshot{
		id:        session.ID(),
		accountID: session.AccountID(),
		tokenHash: session.TokenHash(),
		userAgent: session.UserAgent(),
		ip:        session.IP(),
		expiresAt: session.ExpiresAt(),
		revokedAt: session.RevokedAt(),
		createdAt: session.CreatedAt(),
		updatedAt: session.UpdatedAt(),
	}
}

func cloneSessionForTest(snapshot sessionSnapshot) (*authdomain.RefreshSession, error) {
	return authdomain.RehydrateRefreshSession(authdomain.RehydrateRefreshSessionParams{
		ID:        snapshot.id,
		AccountID: snapshot.accountID,
		TokenHash: snapshot.tokenHash,
		UserAgent: snapshot.userAgent,
		IP:        snapshot.ip,
		ExpiresAt: snapshot.expiresAt,
		RevokedAt: snapshot.revokedAt,
		CreatedAt: snapshot.createdAt,
		UpdatedAt: snapshot.updatedAt,
	})
}

func TestHandlerRotatesRefreshToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	account := testAccount(t, "user@example.com")
	session := testSession(t, account.ID().UUID(), "hash:refresh-old", now)
	repo := newFakeSessionRepo(session)
	handler := NewHandler(
		&fakeAccountReader{account: account},
		repo,
		fakeTokenManager{},
		&fakeRandom{values: []string{"refresh-new"}},
		fakeClock{now: now},
	)

	res, err := handler.Handle(context.Background(), Command{RefreshToken: "refresh-old"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if res.RefreshToken != "refresh-new" {
		t.Fatalf("RefreshToken = %q, want refresh-new", res.RefreshToken)
	}
	if repo.session.TokenHash() != "hash:refresh-new" {
		t.Fatalf("session token hash = %q, want hash:refresh-new", repo.session.TokenHash())
	}
}

func TestHandlerRevokesSessionOnRefreshReuse(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	account := testAccount(t, "reuse@example.com")
	session := testSession(t, account.ID().UUID(), "hash:refresh-old", now)
	repo := newFakeSessionRepo(session)
	handler := NewHandler(
		&fakeAccountReader{account: account},
		repo,
		fakeTokenManager{},
		&fakeRandom{values: []string{"refresh-new", "refresh-unused", "refresh-after-reuse"}},
		fakeClock{now: now},
	)

	first, err := handler.Handle(context.Background(), Command{RefreshToken: "refresh-old"})
	if err != nil {
		t.Fatalf("first Handle() error = %v", err)
	}
	if first.RefreshToken != "refresh-new" {
		t.Fatalf("first RefreshToken = %q, want refresh-new", first.RefreshToken)
	}

	if err := assertRefreshInvalid(handler.Handle(context.Background(), Command{RefreshToken: "refresh-old"})); err != nil {
		t.Fatal(err)
	}
	if !repo.session.IsRevoked() {
		t.Fatal("session must be revoked after refresh token reuse")
	}
	if err := assertRefreshInvalid(handler.Handle(context.Background(), Command{RefreshToken: "refresh-new"})); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerConcurrentRefreshOnlyOneSucceeds(t *testing.T) {
	now := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	account := testAccount(t, "race@example.com")
	session := testSession(t, account.ID().UUID(), "hash:refresh-old", now)
	repo := newFakeSessionRepo(session)
	handler := NewHandler(
		&fakeAccountReader{account: account},
		repo,
		fakeTokenManager{},
		&fakeRandom{values: []string{"refresh-a", "refresh-b"}},
		fakeClock{now: now},
	)

	type outcome struct {
		res *Result
		err error
	}
	results := make(chan outcome, 2)

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := handler.Handle(context.Background(), Command{RefreshToken: "refresh-old"})
			results <- outcome{res: res, err: err}
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for result := range results {
		if result.err == nil {
			successes++
			continue
		}
		if err := assertRefreshInvalid(result.res, result.err); err != nil {
			t.Fatal(err)
		}
		failures++
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("successes=%d failures=%d, want 1/1", successes, failures)
	}
	if !repo.session.IsRevoked() {
		t.Fatal("session must be revoked after concurrent refresh reuse")
	}
}

func testAccount(t *testing.T, email string) *accdomain.Account {
	t.Helper()
	hash, err := accdomain.NewPasswordHash("bcrypt-hash")
	if err != nil {
		t.Fatalf("NewPasswordHash() error = %v", err)
	}
	parsedEmail, err := accdomain.NewEmail(email)
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}
	account, err := accdomain.NewAccount(accdomain.NewAccountParams{
		ID:           accdomain.NewAccountID(),
		Email:        parsedEmail,
		PasswordHash: hash,
		Now:          time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}
	return account
}

func testSession(t *testing.T, accountID uuid.UUID, tokenHash string, now time.Time) *authdomain.RefreshSession {
	t.Helper()
	session, err := authdomain.NewRefreshSession(authdomain.NewRefreshSessionParams{
		ID:        uuid.New(),
		AccountID: accountID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(24 * time.Hour),
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewRefreshSession() error = %v", err)
	}
	return session
}

func assertRefreshInvalid(res *Result, err error) error {
	if err == nil {
		return fault.Validation("expected refresh token invalid error")
	}
	appErr, ok := fault.As(err)
	if !ok {
		return fault.Validation("expected typed fault error")
	}
	if appErr.Kind != fault.KindUnauthorized || appErr.Code != autherrors.CodeRefreshInvalid {
		return fault.Validation("unexpected error kind or code")
	}
	if res != nil {
		return fault.Validation("result must be nil on invalid refresh")
	}
	return nil
}

var (
	_ ports.AccountReader        = (*fakeAccountReader)(nil)
	_ ports.TokenManager         = fakeTokenManager{}
	_ ports.SessionRepository    = (*fakeSessionRepo)(nil)
	_ ports.RandomTokenGenerator = (*fakeRandom)(nil)
	_ ports.Clock                = fakeClock{}
)
