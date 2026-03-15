package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	create := createTenantOp
	create.Middlewares = secured
	huma.Register(api, create, h.CreateTenant)

	listMine := listMyTenantsOp
	listMine.Middlewares = secured
	huma.Register(api, listMine, h.ListMyTenants)

	get := getTenantOp
	get.Middlewares = secured
	huma.Register(api, get, h.GetTenant)

	addMember := addTenantMemberOp
	addMember.Middlewares = secured
	huma.Register(api, addMember, h.AddTenantMember)

	listMembers := listTenantMembersOp
	listMembers.Middlewares = secured
	huma.Register(api, listMembers, h.ListTenantMembers)

	addOrganization := addTenantOrganizationOp
	addOrganization.Middlewares = secured
	huma.Register(api, addOrganization, h.AddTenantOrganization)

	listOrganizations := listTenantOrganizationsOp
	listOrganizations.Middlewares = secured
	huma.Register(api, listOrganizations, h.ListTenantOrganizations)
}
