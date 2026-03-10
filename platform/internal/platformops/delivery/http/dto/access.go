package dto

import "github.com/google/uuid"

type AccessMeInput struct{}

type GetAccountRolesInput struct {
	AccountID string `path:"accountId" required:"true" doc:"Local account id whose platform access should be resolved."`
}

type ReplaceAccountRolesInput struct {
	AccountID string `path:"accountId" required:"true" doc:"Local account id whose stored platform roles should be replaced."`
	Body      struct {
		Roles []string `json:"roles" required:"true" doc:"Stored platform roles to assign. Supported values: platform_admin, support_operator, review_operator."`
	}
}

type PlatformAccessResponse struct {
	Status int `json:"-"`
	Body   struct {
		AccountID      uuid.UUID `json:"accountId"`
		StoredRoles    []string  `json:"storedRoles"`
		EffectiveRoles []string  `json:"effectiveRoles"`
		BootstrapAdmin bool      `json:"bootstrapAdmin"`
	}
}
