package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OIDCAuthorizationRequest struct {
	State  string
	Nonce  string
	Prompt string
}

type OIDCProvider interface {
	Name() string
	BuildAuthorizationURL(ctx context.Context, req OIDCAuthorizationRequest) (string, error)
	ExchangeCode(ctx context.Context, code string) (*OIDCIdentity, error)
}

type OIDCIdentity struct {
	Provider      string
	Subject       string
	Nonce         string
	Email         string
	EmailVerified bool
	DisplayName   *string
	ClaimsJSON    string
}

type ExternalIdentityRecord struct {
	ID              uuid.UUID
	Provider        string
	ExternalSubject string
	AccountID       uuid.UUID
	Email           *string
	EmailVerified   bool
	DisplayName     *string
	ClaimsJSON      string
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

type OAuthStateRecord struct {
	ID        uuid.UUID
	Provider  string
	StateHash string
	ReturnTo  string
	Intent    string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type OIDCNonceRecord struct {
	ID           uuid.UUID
	Provider     string
	OAuthStateID uuid.UUID
	NonceHash    string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type OneTimeCodeRecord struct {
	ID           uuid.UUID
	Purpose      string
	CodeHash     string
	AccountID    uuid.UUID
	Provider     string
	Intent       string
	IsNewAccount bool
	ExpiresAt    time.Time
	UsedAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type ExternalIdentityRepository interface {
	GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*ExternalIdentityRecord, error)
	Create(ctx context.Context, record *ExternalIdentityRecord) error
	TouchLogin(ctx context.Context, id uuid.UUID, email *string, emailVerified bool, displayName *string, claimsJSON string, at time.Time) error
}

type OIDCStateRepository interface {
	CreateState(ctx context.Context, record *OAuthStateRecord) error
	CreateNonce(ctx context.Context, record *OIDCNonceRecord) error
	GetStateByHash(ctx context.Context, provider, stateHash string) (*OAuthStateRecord, error)
	GetNonceByStateID(ctx context.Context, provider string, stateID uuid.UUID) (*OIDCNonceRecord, error)
	MarkStateUsed(ctx context.Context, id uuid.UUID, at time.Time) error
	MarkNonceUsed(ctx context.Context, id uuid.UUID, at time.Time) error
}

type OneTimeCodeRepository interface {
	Create(ctx context.Context, record *OneTimeCodeRecord) error
	GetByCodeHash(ctx context.Context, purpose, codeHash string) (*OneTimeCodeRecord, error)
	MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) (bool, error)
}
