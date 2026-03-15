package access

import (
	"testing"

	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
)

func TestHasOrganizationPermission(t *testing.T) {
	tests := []struct {
		role       memberdomain.MembershipRole
		permission Permission
		want       bool
	}{
		{memberdomain.MembershipRoleOwner, PermissionOrganizationManageMembers, true},
		{memberdomain.MembershipRoleAdmin, PermissionOrganizationManageProfile, true},
		{memberdomain.MembershipRoleManager, PermissionOrganizationManageCatalog, true},
		{memberdomain.MembershipRoleManager, PermissionOrganizationManageMembers, false},
		{memberdomain.MembershipRoleMember, PermissionOrganizationEmployeeAccess, true},
		{memberdomain.MembershipRoleMember, PermissionOrganizationManageCatalog, true},
		{memberdomain.MembershipRoleViewer, PermissionOrganizationEmployeeAccess, false},
		{memberdomain.MembershipRoleViewer, PermissionOrganizationRead, true},
	}

	for _, tt := range tests {
		if got := HasOrganizationPermission(tt.role, tt.permission); got != tt.want {
			t.Fatalf("HasOrganizationPermission(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.want)
		}
	}
}

func TestCanAssignOrganizationRole(t *testing.T) {
	if !CanAssignOrganizationRole(memberdomain.MembershipRoleOwner, memberdomain.MembershipRoleOwner) {
		t.Fatal("owner must be able to assign owner role")
	}
	if CanAssignOrganizationRole(memberdomain.MembershipRoleAdmin, memberdomain.MembershipRoleOwner) {
		t.Fatal("admin must not be able to assign owner role")
	}
	if !CanAssignOrganizationRole(memberdomain.MembershipRoleAdmin, memberdomain.MembershipRoleMember) {
		t.Fatal("admin must be able to assign member role")
	}
}
