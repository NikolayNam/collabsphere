package domain

import (
    "time"

    "github.com/google/uuid"
)

type RefreshSession struct {
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

type NewRefreshSessionParams struct {
    ID        uuid.UUID
    AccountID uuid.UUID
    TokenHash string
    UserAgent *string
    IP        *string
    ExpiresAt time.Time
    Now       time.Time
}

func NewRefreshSession(p NewRefreshSessionParams) (*RefreshSession, error) {
    if p.ID == uuid.Nil {
        return nil, ErrSessionIDRequired
    }
    if p.AccountID == uuid.Nil {
        return nil, ErrAccountIDRequired
    }
    if p.TokenHash == "" {
        return nil, ErrTokenHashRequired
    }
    if p.Now.IsZero() {
        return nil, ErrNowRequired
    }
    if p.ExpiresAt.IsZero() || !p.ExpiresAt.After(p.Now) {
        return nil, ErrSessionExpiresAtInvalid
    }

    return &RefreshSession{
        id:        p.ID,
        accountID: p.AccountID,
        tokenHash: p.TokenHash,
        userAgent: cloneStringPtr(p.UserAgent),
        ip:        cloneStringPtr(p.IP),
        expiresAt: p.ExpiresAt,
        createdAt: p.Now,
    }, nil
}

type RehydrateRefreshSessionParams struct {
    ID        uuid.UUID
    AccountID uuid.UUID
    TokenHash string
    UserAgent *string
    IP        *string
    ExpiresAt time.Time
    RevokedAt *time.Time
    CreatedAt time.Time
    UpdatedAt *time.Time
}

func RehydrateRefreshSession(p RehydrateRefreshSessionParams) (*RefreshSession, error) {
    if p.ID == uuid.Nil {
        return nil, ErrSessionIDRequired
    }
    if p.AccountID == uuid.Nil {
        return nil, ErrAccountIDRequired
    }
    if p.TokenHash == "" {
        return nil, ErrTokenHashRequired
    }
    if p.CreatedAt.IsZero() {
        return nil, ErrNowRequired
    }
    if p.ExpiresAt.IsZero() {
        return nil, ErrSessionExpiresAtInvalid
    }

    return &RefreshSession{
        id:        p.ID,
        accountID: p.AccountID,
        tokenHash: p.TokenHash,
        userAgent: cloneStringPtr(p.UserAgent),
        ip:        cloneStringPtr(p.IP),
        expiresAt: p.ExpiresAt,
        revokedAt: cloneTimePtr(p.RevokedAt),
        createdAt: p.CreatedAt,
        updatedAt: cloneTimePtr(p.UpdatedAt),
    }, nil
}

func (s *RefreshSession) ID() uuid.UUID         { return s.id }
func (s *RefreshSession) AccountID() uuid.UUID  { return s.accountID }
func (s *RefreshSession) TokenHash() string     { return s.tokenHash }
func (s *RefreshSession) UserAgent() *string    { return cloneStringPtr(s.userAgent) }
func (s *RefreshSession) IP() *string           { return cloneStringPtr(s.ip) }
func (s *RefreshSession) ExpiresAt() time.Time  { return s.expiresAt }
func (s *RefreshSession) RevokedAt() *time.Time { return cloneTimePtr(s.revokedAt) }
func (s *RefreshSession) CreatedAt() time.Time  { return s.createdAt }
func (s *RefreshSession) UpdatedAt() *time.Time { return cloneTimePtr(s.updatedAt) }

func (s *RefreshSession) IsRevoked() bool {
    return s.revokedAt != nil
}

func (s *RefreshSession) IsExpired(now time.Time) bool {
    return !s.expiresAt.After(now)
}

func (s *RefreshSession) Revoke(now time.Time) error {
    if now.IsZero() {
        return ErrNowRequired
    }
    if s.revokedAt != nil {
        return nil
    }
    revokedAt := now
    updatedAt := now
    s.revokedAt = &revokedAt
    s.updatedAt = &updatedAt
    return nil
}

func cloneStringPtr(v *string) *string {
    if v == nil {
        return nil
    }
    cloned := *v
    return &cloned
}

func cloneTimePtr(v *time.Time) *time.Time {
    if v == nil {
        return nil
    }
    cloned := *v
    return &cloned
}
