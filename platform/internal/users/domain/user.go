package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrUserInactive     = errors.New("user is inactive")
	ErrInvalidFirstName = errors.New("first name is invalid")
	ErrInvalidLastName  = errors.New("last name is invalid")
	ErrInvalidPhone     = errors.New("phone is invalid")
	ErrCannotDeactivate = errors.New("cannot deactivate user")
	ErrCannotActivate   = errors.New("cannot activate user")
)

type User struct {
	// identity
	id    UserID
	email Email

	// credentials (секрет)
	passwordHash PasswordHash

	// profile
	firstName string
	lastName  string
	phone     string

	// access
	isActive bool

	// Domain time (только если нужно домену; иначе убрать)
	createdAt time.Time
	updatedAt time.Time
}

type NewUserParams struct {
	ID           UserID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Phone        string
	Now          time.Time
}

func NewUser(p NewUserParams) (*User, error) {
	if err := validateUserCore(p.ID, p.Email, p.PasswordHash); err != nil {
		return nil, err
	}
	if p.Now.IsZero() {
		return nil, errors.New("now is required")
	}

	fn, ln, ph, err := normalizeProfile(p.FirstName, p.LastName, p.Phone)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		phone:        ph,
		isActive:     true,
		createdAt:    p.Now,
		updatedAt:    p.Now,
	}, nil
}

// RehydrateUserParams — восстановление из persistence (репозиторий).
type RehydrateUserParams struct {
	ID           UserID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Phone        string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func RehydrateUser(p RehydrateUserParams) (*User, error) {
	if err := validateUserCore(p.ID, p.Email, p.PasswordHash); err != nil {
		return nil, err
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, errors.New("timestamps are required")
	}

	fn, ln, ph, err := normalizeProfile(p.FirstName, p.LastName, p.Phone)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		phone:        ph,
		isActive:     p.IsActive,
		createdAt:    p.CreatedAt,
		updatedAt:    p.UpdatedAt,
	}, nil
}

/*
	Read-only API (наружу — только безопасное)
*/

func (u *User) ID() UserID     { return u.id }
func (u *User) Email() Email   { return u.email }
func (u *User) IsActive() bool { return u.isActive }

func (u *User) FirstName() string { return u.firstName }
func (u *User) LastName() string  { return u.lastName }
func (u *User) Phone() string     { return u.phone }

func (u *User) PasswordHash() PasswordHash { return u.passwordHash }

func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

/*
	Domain behavior
*/

func (u *User) Activate(now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if u.isActive {
		return ErrCannotActivate
	}
	u.isActive = true
	u.updatedAt = now
	return nil
}

func (u *User) Deactivate(now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if !u.isActive {
		return ErrCannotDeactivate
	}
	u.isActive = false
	u.updatedAt = now
	return nil
}

func (u *User) UpdateProfile(firstName, lastName, phone string, now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if !u.isActive {
		return ErrUserInactive
	}

	fn, ln, ph, err := normalizeProfile(firstName, lastName, phone)
	if err != nil {
		return err
	}

	u.firstName = fn
	u.lastName = ln
	u.phone = ph
	u.updatedAt = now
	return nil
}

func (u *User) SetPasswordHash(hash PasswordHash, now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if hash.IsZero() {
		return errors.New("password hash is empty")
	}
	if !u.isActive {
		return ErrUserInactive
	}

	u.passwordHash = hash
	u.updatedAt = now
	return nil
}

/*
	internal validation helpers
*/

func validateUserCore(id UserID, email Email, hash PasswordHash) error {
	if id.IsZero() {
		return errors.New("missing user id")
	}
	if email.IsZero() {
		return errors.New("missing email")
	}
	if hash.IsZero() {
		return errors.New("missing password hash")
	}
	return nil
}

func normalizeProfile(firstName, lastName, phone string) (string, string, string, error) {
	fn, err := normalizeFirstName(firstName)
	if err != nil {
		return "", "", "", err
	}
	ln, err := normalizeLastName(lastName)
	if err != nil {
		return "", "", "", err
	}
	ph, err := normalizePhone(phone)
	if err != nil {
		return "", "", "", err
	}
	return fn, ln, ph, nil
}

func normalizeFirstName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if len(s) > 200 {
		return "", ErrInvalidFirstName
	}
	return s, nil
}

func normalizeLastName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if len(s) > 200 {
		return "", ErrInvalidLastName
	}
	return s, nil
}

func normalizePhone(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > 50 {
		return "", ErrInvalidPhone
	}
	return s, nil
}
