package mapper

import (
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/groups/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

func ToGroupResponse(group *domain.Group, status int) *dto.GroupResponse {
	if group == nil {
		return nil
	}
	return &dto.GroupResponse{
		Status: status,
		Body: dto.GroupBody{
			ID:          group.ID().UUID(),
			Name:        group.Name(),
			Slug:        group.Slug(),
			Description: group.Description(),
			IsActive:    group.IsActive(),
			CreatedAt:   group.CreatedAt(),
		},
	}
}

func ToAccountMemberResponse(member domain.AccountMemberView) *dto.GroupAccountMemberResponse {
	return &dto.GroupAccountMemberResponse{
		Status: http.StatusCreated,
		Body: dto.GroupAccountMemberBody{
			ID:          member.MembershipID,
			AccountID:   member.AccountID,
			Email:       member.Email,
			DisplayName: member.DisplayName,
			Role:        member.Role,
			IsActive:    member.IsActive,
			CreatedAt:   member.CreatedAt,
		},
	}
}

func ToOrganizationMemberResponse(member domain.OrganizationMemberView) *dto.GroupOrganizationMemberResponse {
	return &dto.GroupOrganizationMemberResponse{
		Status: http.StatusCreated,
		Body: dto.GroupOrganizationMemberBody{
			ID:             member.MembershipID,
			OrganizationID: member.OrganizationID,
			Name:           member.Name,
			Slug:           member.Slug,
			IsActive:       member.IsActive,
			CreatedAt:      member.CreatedAt,
		},
	}
}

func ToMembersResponse(members *domain.MembersView) *dto.GroupMembersResponse {
	response := &dto.GroupMembersResponse{
		Status: http.StatusOK,
		Body: dto.GroupMembersBody{
			Accounts:      []dto.GroupAccountMemberBody{},
			Organizations: []dto.GroupOrganizationMemberBody{},
		},
	}
	if members == nil {
		return response
	}

	for _, member := range members.Accounts {
		response.Body.Accounts = append(response.Body.Accounts, dto.GroupAccountMemberBody{
			ID:          member.MembershipID,
			AccountID:   member.AccountID,
			Email:       member.Email,
			DisplayName: member.DisplayName,
			Role:        member.Role,
			IsActive:    member.IsActive,
			CreatedAt:   member.CreatedAt,
		})
	}
	for _, member := range members.Organizations {
		response.Body.Organizations = append(response.Body.Organizations, dto.GroupOrganizationMemberBody{
			ID:             member.MembershipID,
			OrganizationID: member.OrganizationID,
			Name:           member.Name,
			Slug:           member.Slug,
			IsActive:       member.IsActive,
			CreatedAt:      member.CreatedAt,
		})
	}
	return response
}

func ToMyGroupsResponse(items []dto.MyGroupBody) *dto.MyGroupsResponse {
	response := &dto.MyGroupsResponse{Status: http.StatusOK}
	response.Body.Data = make([]dto.MyGroupBody, 0, len(items))
	response.Body.Data = append(response.Body.Data, items...)
	return response
}
