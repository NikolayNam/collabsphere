package domain

import (
    "strings"
    "time"
    "unicode/utf8"
)

type Account struct {
    id           AccountID
    email        Email
    passwordHash PasswordHash
    displayName  *string
    isActive     bool
    createdAt    time.Time
    updatedAt    *time.Time
}

type NewAccountParams struct {
    ID           AccountID
    Email        Email
    PasswordHash PasswordHash
    DisplayName  *string
    Now          time.Time
}

func NewAccount(p NewAccountParams) (*Account, error) {
    if err := validateAccountCore(p.ID, p.Email, p.PasswordHash); err != nil {
        return nil, err
    }
    if p.Now.IsZero() {
        return nil, ErrNowRequired
    }

    displayName, err := normalizeOptionalDisplayName(p.DisplayName)
    if err != nil {
        return nil, err
    }

    updatedAt := p.Now

    return &Account{
        id:           p.ID,
        email:        p.Email,
        passwordHash: p.PasswordHash,
        displayName:  displayName,
        isActive:     true,
        createdAt:    p.Now,
        updatedAt:    &updatedAt,
    }, nil
}

type RehydrateAccountParams struct {
    ID           AccountID
    Email        Email
    PasswordHash PasswordHash
    DisplayName  *string
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func RehydrateAccount(p RehydrateAccountParams) (*Account, error) {
    if err := validateAccountCore(p.ID, p.Email, p.PasswordHash); err != nil {
        return nil, err
    }
    if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
        return nil, ErrTimestampsMissing
    }
    if p.UpdatedAt.Before(p.CreatedAt) {
        return nil, ErrTimestampsInvalid
    }

    displayName, err := normalizeOptionalDisplayName(p.DisplayName)
    if err != nil {
        return nil, err
    }

    updatedAt := p.UpdatedAt

    return &Account{
        id:           p.ID,
        email:        p.Email,
        passwordHash: p.PasswordHash,
        displayName:  displayName,
        isActive:     p.IsActive,
        createdAt:    p.CreatedAt,
        updatedAt:    &updatedAt,
    }, nil
}

func (a *Account) ID() AccountID {
    return a.id
}

func (a *Account) Email() Email {
    return a.email
}

func (a *Account) PasswordHash() PasswordHash {
    return a.passwordHash
}

func (a *Account) DisplayName() *string {
    return cloneStringPtr(a.displayName)
}

func (a *Account) IsActive() bool {
    return a.isActive
}

func (a *Account) Status() AccountStatus {
    if a.isActive {
        return AccountStatusActive
    }
    return AccountStatusBlocked
}

func (a *Account) CreatedAt() time.Time {
    return a.createdAt
}

func (a *Account) UpdatedAt() *time.Time {
    return cloneTimePtr(a.updatedAt)
}

func validateAccountCore(id AccountID, email Email, hash PasswordHash) error {
    switch {
    case id.IsZero():
        return ErrUserIDEmpty
    case email.IsZero():
        return ErrEmailEmpty
    case hash.IsZero():
        return ErrPasswordHashEmpty
    default:
        return nil
    }
}

func normalizeOptionalDisplayName(s *string) (*string, error) {
    if s == nil {
        return nil, nil
    }

    v := strings.TrimSpace(*s)
    if v == "" {
        return nil, nil
    }
    if utf8.RuneCountInString(v) > 255 {
        return nil, ErrDisplayNameInvalid
    }

    return &v, nil
}

func cloneTimePtr(t *time.Time) *time.Time {
    if t == nil {
        return nil
    }
    v := *t
    return &v
}

func cloneStringPtr(s *string) *string {
    if s == nil {
        return nil
    }
    v := *s
    return &v
}
