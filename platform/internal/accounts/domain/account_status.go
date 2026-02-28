package domain

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusBlocked   AccountStatus = "blocked"
)

func (s AccountStatus) IsValid() bool {
	switch s {
	case AccountStatusActive, AccountStatusSuspended, AccountStatusBlocked:
		return true
	default:
		return false
	}
}

func NewAccountStatus(v string) (AccountStatus, error) {
	s := AccountStatus(v)
	if !s.IsValid() {
		return "", ErrInvalidAccountStatus
	}
	return s, nil
}
