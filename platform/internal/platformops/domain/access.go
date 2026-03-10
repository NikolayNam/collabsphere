package domain

import "github.com/google/uuid"

type Access struct {
	AccountID      uuid.UUID
	StoredRoles    []Role
	EffectiveRoles []Role
	BootstrapAdmin bool
}

func (a Access) HasAnyRole(required ...Role) bool {
	if len(required) == 0 {
		return len(a.EffectiveRoles) > 0
	}
	have := make(map[Role]struct{}, len(a.EffectiveRoles))
	for _, role := range a.EffectiveRoles {
		have[role] = struct{}{}
	}
	for _, role := range required {
		if _, ok := have[role]; ok {
			return true
		}
	}
	return false
}
