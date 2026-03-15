package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	org "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type OrganizationRole struct {
	id             uuid.UUID
	organizationID org.OrganizationID
	code           string
	name           string
	description    string
	baseRole       MembershipRole
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

type NewOrganizationRoleParams struct {
	ID             uuid.UUID
	OrganizationID org.OrganizationID
	Code           string
	Name           string
	Description    string
	BaseRole       string
	Now            time.Time
}

type RehydrateOrganizationRoleParams struct {
	ID             uuid.UUID
	OrganizationID org.OrganizationID
	Code           string
	Name           string
	Description    string
	BaseRole       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type OrganizationRolePatch struct {
	Name        *string
	Description *string
	BaseRole    *string
	UpdatedAt   time.Time
}

func NewOrganizationRole(p NewOrganizationRoleParams) (*OrganizationRole, error) {
	if p.ID == uuid.Nil || p.OrganizationID.IsZero() || p.Now.IsZero() {
		return nil, ErrOrganizationRoleInvalid
	}
	code, err := normalizeRoleCode(p.Code)
	if err != nil {
		return nil, err
	}
	name, err := normalizeRoleName(p.Name)
	if err != nil {
		return nil, err
	}
	baseRole, err := parseBaseRole(p.BaseRole)
	if err != nil {
		return nil, err
	}
	if isSystemRoleCode(code) {
		return nil, ErrOrganizationRoleCodeReserved
	}
	now := p.Now
	return &OrganizationRole{
		id:             p.ID,
		organizationID: p.OrganizationID,
		code:           code,
		name:           name,
		description:    strings.TrimSpace(p.Description),
		baseRole:       baseRole,
		createdAt:      now,
		updatedAt:      now,
		deletedAt:      nil,
	}, nil
}

func RehydrateOrganizationRole(p RehydrateOrganizationRoleParams) (*OrganizationRole, error) {
	if p.ID == uuid.Nil || p.OrganizationID.IsZero() {
		return nil, ErrOrganizationRoleInvalid
	}
	code, err := normalizeRoleCode(p.Code)
	if err != nil {
		return nil, err
	}
	name, err := normalizeRoleName(p.Name)
	if err != nil {
		return nil, err
	}
	baseRole, err := parseBaseRole(p.BaseRole)
	if err != nil {
		return nil, err
	}
	return &OrganizationRole{
		id:             p.ID,
		organizationID: p.OrganizationID,
		code:           code,
		name:           name,
		description:    strings.TrimSpace(p.Description),
		baseRole:       baseRole,
		createdAt:      p.CreatedAt,
		updatedAt:      p.UpdatedAt,
		deletedAt:      cloneTimePtr(p.DeletedAt),
	}, nil
}

func (r *OrganizationRole) ID() uuid.UUID                    { return r.id }
func (r *OrganizationRole) OrganizationID() org.OrganizationID { return r.organizationID }
func (r *OrganizationRole) Code() string                     { return r.code }
func (r *OrganizationRole) Name() string                     { return r.name }
func (r *OrganizationRole) Description() string             { return r.description }
func (r *OrganizationRole) BaseRole() MembershipRole        { return r.baseRole }
func (r *OrganizationRole) CreatedAt() time.Time             { return r.createdAt }
func (r *OrganizationRole) UpdatedAt() time.Time             { return r.updatedAt }
func (r *OrganizationRole) DeletedAt() *time.Time            { return cloneTimePtr(r.deletedAt) }
func (r *OrganizationRole) IsDeleted() bool                  { return r.deletedAt != nil }

func (r *OrganizationRole) ApplyPatch(p OrganizationRolePatch) error {
	if r == nil {
		return ErrOrganizationRoleInvalid
	}
	if p.UpdatedAt.IsZero() {
		return ErrNowRequired
	}
	if p.Name != nil {
		name, err := normalizeRoleName(*p.Name)
		if err != nil {
			return err
		}
		r.name = name
	}
	if p.Description != nil {
		r.description = strings.TrimSpace(*p.Description)
	}
	if p.BaseRole != nil {
		baseRole, err := parseBaseRole(*p.BaseRole)
		if err != nil {
			return err
		}
		r.baseRole = baseRole
	}
	r.updatedAt = p.UpdatedAt
	return nil
}

func (r *OrganizationRole) SoftDelete(now time.Time) error {
	if r == nil || now.IsZero() {
		return ErrOrganizationRoleInvalid
	}
	r.deletedAt = &now
	r.updatedAt = now
	return nil
}

func normalizeRoleCode(s string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if v == "" || utf8.RuneCountInString(v) > 64 {
		return "", ErrOrganizationRoleCodeInvalid
	}
	if !codePattern.MatchString(v) {
		return "", ErrOrganizationRoleCodeInvalid
	}
	return v, nil
}

func normalizeRoleName(s string) (string, error) {
	v := strings.TrimSpace(s)
	if v == "" || utf8.RuneCountInString(v) > 255 {
		return "", ErrOrganizationRoleNameInvalid
	}
	return v, nil
}

func parseBaseRole(s string) (MembershipRole, error) {
	r := MembershipRole(strings.ToLower(strings.TrimSpace(s)))
	if !r.IsValid() {
		return "", ErrOrganizationRoleBaseRoleInvalid
	}
	return r, nil
}

func isSystemRoleCode(code string) bool {
	switch code {
	case "owner", "admin", "manager", "member", "viewer":
		return true
	default:
		return false
	}
}
