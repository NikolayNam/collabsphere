package domain

import (
	"time"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type OrganizationMember struct {
	id             uuid.UUID
	groupID        GroupID
	organizationID orgdomain.OrganizationID
	isActive       bool
	createdAt      time.Time
	updatedAt      *time.Time
}

type NewOrganizationMemberParams struct {
	GroupID        GroupID
	OrganizationID orgdomain.OrganizationID
	Now            time.Time
}

type RehydrateOrganizationMemberParams struct {
	ID             uuid.UUID
	GroupID        GroupID
	OrganizationID orgdomain.OrganizationID
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewOrganizationMember(p NewOrganizationMemberParams) (*OrganizationMember, error) {
	if p.GroupID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrOrgMemberInvalid
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	updatedAt := p.Now
	return &OrganizationMember{
		id:             uuid.New(),
		groupID:        p.GroupID,
		organizationID: p.OrganizationID,
		isActive:       true,
		createdAt:      p.Now,
		updatedAt:      &updatedAt,
	}, nil
}

func RehydrateOrganizationMember(p RehydrateOrganizationMemberParams) (*OrganizationMember, error) {
	if p.ID == uuid.Nil || p.GroupID.IsZero() || p.OrganizationID.IsZero() {
		return nil, ErrOrgMemberInvalid
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	updatedAt := p.UpdatedAt
	return &OrganizationMember{
		id:             p.ID,
		groupID:        p.GroupID,
		organizationID: p.OrganizationID,
		isActive:       p.IsActive,
		createdAt:      p.CreatedAt,
		updatedAt:      &updatedAt,
	}, nil
}

func (m *OrganizationMember) ID() uuid.UUID                            { return m.id }
func (m *OrganizationMember) GroupID() GroupID                         { return m.groupID }
func (m *OrganizationMember) OrganizationID() orgdomain.OrganizationID { return m.organizationID }
func (m *OrganizationMember) IsActive() bool                           { return m.isActive }
func (m *OrganizationMember) CreatedAt() time.Time                     { return m.createdAt }
func (m *OrganizationMember) UpdatedAt() *time.Time                    { return cloneTimePtr(m.updatedAt) }
