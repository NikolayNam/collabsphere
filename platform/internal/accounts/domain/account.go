package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Account struct {
	// identity
	id    AccountID
	email Email

	// credentials
	passwordHash PasswordHash

	// profile
	firstName string
	lastName  string

	// status
	status AccountStatus

	// timestamps
	createdAt time.Time
	updatedAt *time.Time
}

type NewAccountParams struct {
	ID           AccountID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Now          time.Time
}

func NewAccount(p NewAccountParams) (*Account, error) {
	if err := validateAccountCore(p.ID, p.Email, p.PasswordHash); err != nil {
		return nil, err
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	fn, ln, err := normalizeAccountNames(p.FirstName, p.LastName)
	if err != nil {
		return nil, err
	}

	return &Account{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		status:       AccountStatusActive,
		createdAt:    p.Now,
		updatedAt:    nil,
	}, nil
}

type RehydrateAccountParams struct {
	ID           AccountID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Status       AccountStatus
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func RehydrateAccount(p RehydrateAccountParams) (*Account, error) {
	if err := validateAccountCore(p.ID, p.Email, p.PasswordHash); err != nil {
		return nil, err
	}
	if p.CreatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt != nil && p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	fn, ln, err := normalizeAccountNames(p.FirstName, p.LastName)
	if err != nil {
		return nil, err
	}

	return &Account{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		status:       p.Status,
		createdAt:    p.CreatedAt,
		updatedAt:    cloneTimePtr(p.UpdatedAt),
	}, nil
}

// Accessors

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) Email() Email {
	return a.email
}

func (a *Account) PasswordHash() PasswordHash {
	return a.passwordHash
}

func (a *Account) FirstName() string {
	return a.firstName
}

func (a *Account) LastName() string {
	return a.lastName
}

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Account) UpdatedAt() *time.Time {
	return cloneTimePtr(a.updatedAt)
}

// Behavior

func (a *Account) ensureMutable() error {
	switch a.status {
	case AccountStatusActive:
		return nil
	case AccountStatusSuspended:
		return ErrAccountSuspended
	case AccountStatusBlocked:
		return ErrAccountBlocked
	default:
		return ErrInvalidAccountStatus
	}
}

func (a *Account) Activate(now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}
	if err := a.ensureMutable(); err != nil {
		return err
	}

	a.status = AccountStatusActive
	a.touch(now)
	return nil
}

func (a *Account) Block(now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}

	switch a.status {
	case AccountStatusBlocked:
		return ErrAccountBlocked
	case AccountStatusSuspended:
		return ErrAccountStateTransitionNotAllowed
	case AccountStatusActive:
		a.status = AccountStatusBlocked
		a.touch(now)
		return nil
	default:
		return ErrInvalidAccountStatus
	}

}

func (a *Account) UpdateProfile(firstName, lastName string, now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}

	if err := a.ensureMutable(); err != nil {
		return err
	}

	fn, ln, err := normalizeAccountNames(firstName, lastName)
	if err != nil {
		return err
	}

	if a.firstName == fn && a.lastName == ln {
		return nil
	}

	a.firstName = fn
	a.lastName = ln
	a.touch(now)
	return nil
}

func (a *Account) SetPasswordHash(hash PasswordHash, now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}
	if hash.IsZero() {
		return ErrPasswordHashEmpty
	}

	switch a.status {
	case AccountStatusBlocked:
		return ErrAccountBlocked
	case AccountStatusActive:
		a.passwordHash = hash
		a.touch(now)
		return nil
	default:
		return ErrInvalidAccountStatus
	}
}

// Internal validation helpers
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

func normalizeAccountNames(firstName, lastName string) (string, string, error) {
	fn, err := normalizeFirstName(firstName)
	if err != nil {
		return "", lastName, err
	}

	ln, err := normalizeLastName(lastName)
	if err != nil {
		return firstName, "", err
	}

	return fn, ln, nil
}

func normalizeFirstName(s string) (string, error) {
	return normalizeRequiredHumanName(s, ErrInvalidFirstName)
}

func normalizeLastName(s string) (string, error) {
	return normalizeRequiredHumanName(s, ErrInvalidLastName)
}

func normalizeRequiredHumanName(s string, errInvalid error) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > 200 {
		return "", errInvalid
	}
	return s, nil
}

func (a *Account) touch(now time.Time) {
	a.updatedAt = new(now)
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return new(*t)
}
