package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	addMember := addMemberOp
	addMember.Middlewares = secured
	huma.Register(api, addMember, h.AddMember)

	listMembers := listMembersOp
	listMembers.Middlewares = secured
	huma.Register(api, listMembers, h.ListMembers)

	updateMember := updateMemberOp
	updateMember.Middlewares = secured
	huma.Register(api, updateMember, h.UpdateMember)

	removeMember := removeMemberOp
	removeMember.Middlewares = secured
	huma.Register(api, removeMember, h.RemoveMember)

	createInvitation := createInvitationOp
	createInvitation.Middlewares = secured
	huma.Register(api, createInvitation, h.CreateInvitation)

	listInvitations := listInvitationsOp
	listInvitations.Middlewares = secured
	huma.Register(api, listInvitations, h.ListInvitations)

	acceptInvitation := acceptInvitationOp
	acceptInvitation.Middlewares = secured
	huma.Register(api, acceptInvitation, h.AcceptInvitation)
}
