package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	create := createGroupOp
	create.Middlewares = secured
	huma.Register(api, create, h.CreateGroup)

	listMy := listMyGroupsOp
	listMy.Middlewares = secured
	huma.Register(api, listMy, h.ListMyGroups)

	get := getGroupByIDOp
	get.Middlewares = secured
	huma.Register(api, get, h.GetGroupByID)

	addAccount := addAccountMemberOp
	addAccount.Middlewares = secured
	huma.Register(api, addAccount, h.AddAccountMember)

	addOrganization := addOrganizationMemberOp
	addOrganization.Middlewares = secured
	huma.Register(api, addOrganization, h.AddOrganizationMember)

	list := listMembersOp
	list.Middlewares = secured
	huma.Register(api, list, h.ListMembers)
}
