package access

import memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"

type Permission string

const (
	PermissionOrganizationRead             Permission = "organization.read"
	PermissionOrganizationEmployeeAccess   Permission = "organization.employee_access"
	PermissionOrganizationManageProfile    Permission = "organization.manage_profile"
	PermissionOrganizationManageMembers    Permission = "organization.manage_members"
	PermissionOrganizationManageCatalog    Permission = "organization.manage_catalog"
	PermissionOrganizationManagePayments   Permission = "organization.manage_payments"
	PermissionOrganizationManageOnboarding Permission = "organization.manage_onboarding"
)

func HasOrganizationPermission(role memberdomain.MembershipRole, permission Permission) bool {
	switch permission {
	case PermissionOrganizationRead:
		return role.IsValid()
	case PermissionOrganizationEmployeeAccess:
		return role == memberdomain.MembershipRoleOwner ||
			role == memberdomain.MembershipRoleAdmin ||
			role == memberdomain.MembershipRoleManager ||
			role == memberdomain.MembershipRoleMember
	case PermissionOrganizationManageProfile, PermissionOrganizationManageMembers, PermissionOrganizationManageOnboarding:
		return role == memberdomain.MembershipRoleOwner || role == memberdomain.MembershipRoleAdmin
	case PermissionOrganizationManageCatalog, PermissionOrganizationManagePayments:
		return role == memberdomain.MembershipRoleOwner ||
			role == memberdomain.MembershipRoleAdmin ||
			role == memberdomain.MembershipRoleManager
	default:
		return false
	}
}

func CanAssignOrganizationRole(actorRole, targetRole memberdomain.MembershipRole) bool {
	switch actorRole {
	case memberdomain.MembershipRoleOwner:
		return targetRole.IsValid()
	case memberdomain.MembershipRoleAdmin:
		return targetRole == memberdomain.MembershipRoleManager ||
			targetRole == memberdomain.MembershipRoleMember ||
			targetRole == memberdomain.MembershipRoleViewer
	default:
		return false
	}
}

func CanManageOrganizationRole(actorRole, targetRole memberdomain.MembershipRole) bool {
	switch actorRole {
	case memberdomain.MembershipRoleOwner:
		return targetRole.IsValid()
	case memberdomain.MembershipRoleAdmin:
		return targetRole == memberdomain.MembershipRoleManager ||
			targetRole == memberdomain.MembershipRoleMember ||
			targetRole == memberdomain.MembershipRoleViewer
	default:
		return false
	}
}
