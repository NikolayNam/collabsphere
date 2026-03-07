package http

import (
	"context"
	"errors"
	"net/http"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/memberships/application"
	"github.com/NikolayNam/collabsphere/internal/memberships/delivery/http/dto"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) AddMember(ctx context.Context, input *dto.AddMemberInput) (*dto.MembersResponse, error) {
	orgUUID, err := uuid.Parse(input.Body.OrganizationID)
	if err != nil || orgUUID == uuid.Nil {
		return nil, humaerr.From(ctx, application.ErrValidation)
	}
	orgID, _ := orgDomain.OrganizationIDFromUUID(orgUUID)

	accUUID, err := uuid.Parse(input.Body.AccountID)
	if err != nil || accUUID == uuid.Nil {
		return nil, humaerr.From(ctx, application.ErrValidation)
	}

	if err := h.svc.AddMember(ctx, orgID, input.Body.AccountID, input.Body.Kind); err != nil {
		return nil, humaerr.From(ctx, err)
	}

	// Вытащим созданного участника, чтобы вернуть MemberBody (svc.AddMember не возвращает entity/view).
	members, err := h.svc.ListMembers(ctx, orgID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	var created *memberDomain.MemberView
	for i := range members {
		if members[i].AccountID == accUUID {
			created = &members[i]
			break
		}
	}
	if created == nil {
		// Теоретически не должно случиться, если AddMember прошёл успешно.
		return nil, humaerr.From(ctx, errors.New("member not found after successful add"))
	}

	return &dto.MembersResponse{
		OrganizationID: orgUUID,
		Status:         http.StatusCreated,
		Body: dto.MemberBody{
			ID:        created.MembershipID,
			AccountID: created.AccountID,
			Kind:      created.Kind,
			Status:    created.Status,
			CreatedAt: created.CreatedAt,
		},
	}, nil
}

func (h *Handler) ListMembers(ctx context.Context, input *dto.ListMembersInput) (*dto.MembersListResponse, error) {
	orgUUID, err := uuid.Parse(input.OrganizationID)
	if err != nil || orgUUID == uuid.Nil {
		return nil, humaerr.From(ctx, application.ErrValidation)
	}
	orgID, _ := orgDomain.OrganizationIDFromUUID(orgUUID)

	members, err := h.svc.ListMembers(ctx, orgID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	out := &dto.MembersListResponse{
		Status: http.StatusOK,
		Body: dto.MembersListBody{
			Data: make([]dto.MemberBody, 0, len(members)),
		},
	}

	for _, m := range members {
		out.Body.Data = append(out.Body.Data, dto.MemberBody{
			ID:        m.MembershipID,
			AccountID: m.AccountID,
			Kind:      m.Kind,
			Status:    m.Status,
			CreatedAt: m.CreatedAt,
		})
	}

	return out, nil
}
