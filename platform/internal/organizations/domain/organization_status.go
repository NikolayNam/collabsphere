package domain

type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

func (s OrganizationStatus) IsValid() bool {
	switch s {
	case OrganizationStatusActive, OrganizationStatusSuspended, OrganizationStatusArchived:
		return true
	default:
		return false
	}
}

func NewOrganizationStatus(v string) (OrganizationStatus, error) {
	s := OrganizationStatus(v)
	if !s.IsValid() {
		return "", ErrInvalidOrganizationStatus
	}
	return s, nil
}
