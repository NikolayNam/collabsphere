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

	createAccessRequest := createAccessRequestOp
	createAccessRequest.Middlewares = secured
	huma.Register(api, createAccessRequest, h.CreateAccessRequest)

	listAccessRequests := listAccessRequestsOp
	listAccessRequests.Middlewares = secured
	huma.Register(api, listAccessRequests, h.ListAccessRequests)

	approveAccessRequest := approveAccessRequestOp
	approveAccessRequest.Middlewares = secured
	huma.Register(api, approveAccessRequest, h.ApproveAccessRequest)

	rejectAccessRequest := rejectAccessRequestOp
	rejectAccessRequest.Middlewares = secured
	huma.Register(api, rejectAccessRequest, h.RejectAccessRequest)

	listOrganizationRoles := listOrganizationRolesOp
	listOrganizationRoles.Middlewares = secured
	huma.Register(api, listOrganizationRoles, h.ListOrganizationRoles)

	createOrganizationRole := createOrganizationRoleOp
	createOrganizationRole.Middlewares = secured
	huma.Register(api, createOrganizationRole, h.CreateOrganizationRole)

	updateOrganizationRole := updateOrganizationRoleOp
	updateOrganizationRole.Middlewares = secured
	huma.Register(api, updateOrganizationRole, h.UpdateOrganizationRole)

	deleteOrganizationRole := deleteOrganizationRoleOp
	deleteOrganizationRole.Middlewares = secured
	huma.Register(api, deleteOrganizationRole, h.DeleteOrganizationRole)
}
