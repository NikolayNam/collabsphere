package domain

import (
	"sort"
	"strings"
)

type Role string

const (
	RolePlatformAdmin   Role = "platform_admin"
	RoleSupportOperator Role = "support_operator"
	RoleReviewOperator  Role = "review_operator"
)

func ParseRole(raw string) Role {
	return Role(strings.ToLower(strings.TrimSpace(raw)))
}

func (r Role) IsValid() bool {
	switch r {
	case RolePlatformAdmin, RoleSupportOperator, RoleReviewOperator:
		return true
	default:
		return false
	}
}

func (r Role) CanManageAccess() bool {
	return r == RolePlatformAdmin
}

func (r Role) CanForceVerifyZitadelUsers() bool {
	return r == RolePlatformAdmin
}

func (r Role) CanViewUploadQueue() bool {
	return r == RolePlatformAdmin || r == RoleSupportOperator
}

func (r Role) CanViewDashboardSummary() bool {
	return r.IsValid()
}

func RoleStrings(roles []Role) []string {
	out := make([]string, 0, len(roles))
	for _, role := range UniqueSortedRoles(roles) {
		out = append(out, string(role))
	}
	return out
}

func UniqueSortedRoles(values []Role) []Role {
	seen := make(map[Role]struct{}, len(values))
	out := make([]Role, 0, len(values))
	for _, value := range values {
		if !value.IsValid() {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		return roleOrder(out[i]) < roleOrder(out[j])
	})
	return out
}

func roleOrder(role Role) int {
	switch role {
	case RolePlatformAdmin:
		return 0
	case RoleSupportOperator:
		return 1
	case RoleReviewOperator:
		return 2
	default:
		return 99
	}
}
