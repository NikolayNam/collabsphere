package get_organization_by_id

import (
	"context"
	"strings"

	"github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type Handler struct {
	repo ports.OrganizationRepository
}

func NewHandler(repo ports.OrganizationRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*domain.Organization, error) {
	raw := strings.TrimSpace(q.ID)
	if raw == "" {
		return nil, errors.InvalidInput("Invalid organization id")
	}

	uid, err := uuid.Parse(raw)
	if err != nil || uid == uuid.Nil {
		return nil, errors.InvalidInput("Invalid organization id")
	}

	id, err := domain.OrganizationIDFromUUID(uid)
	if err != nil {
		return nil, errors.InvalidInput("Invalid organization id")
	}

	t, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.OrganizationNotFound()
	}

	return t, nil
}
