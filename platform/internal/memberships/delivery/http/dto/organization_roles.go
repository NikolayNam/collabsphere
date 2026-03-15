package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListOrganizationRolesInput struct {
	OrganizationID  string `path:"organization_id" required:"true" format:"uuid"`
	IncludeDeleted  bool   `query:"includeDeleted" doc:"Include soft-deleted roles"`
}

type CreateOrganizationRoleInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	Body           struct {
		Code        string `json:"code" required:"true" maxLength:"64" doc:"Unique role code (lowercase, alphanumeric + underscore)"`
		Name        string `json:"name" required:"true" maxLength:"255"`
		Description string `json:"description,omitempty"`
		BaseRole    string `json:"baseRole" required:"true" doc:"System role to extend: owner, admin, manager, member, viewer"`
	}
}

type UpdateOrganizationRoleInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	RoleID         string `path:"role_id" required:"true" format:"uuid"`
	Body           struct {
		Name        *string `json:"name,omitempty" maxLength:"255"`
		Description *string `json:"description,omitempty"`
		BaseRole    *string `json:"baseRole,omitempty" doc:"System role: owner, admin, manager, member, viewer"`
	}
}

type DeleteOrganizationRoleInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	RoleID         string `path:"role_id" required:"true" format:"uuid"`
}

type OrganizationRolePayload struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	BaseRole       string     `json:"baseRole"`
	IsSystem       bool       `json:"isSystem" doc:"True for built-in roles (owner, admin, etc.)"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type OrganizationRoleResponse struct {
	Status int
	Body   OrganizationRolePayload
}

type OrganizationRolesListResponse struct {
	Status int
	Body   struct {
		Roles []OrganizationRolePayload `json:"roles"`
	}
}
