package create_organization

import (
    "context"
    stdErrors "errors"

    "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
    "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
    "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Handler struct {
    repo  ports.OrganizationRepository
    clock ports.Clock
}

func NewHandler(repo ports.OrganizationRepository, clock ports.Clock) *Handler {
    return &Handler{repo: repo, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*domain.Organization, error) {
    t, err := domain.NewOrganization(domain.NewOrganizationParams{
        ID:   domain.NewOrganizationID(),
        Name: cmd.Name,
        Slug: cmd.Slug,
        Now:  h.clock.Now(),
    })
    if err != nil {
        return nil, errors.InvalidInput("Invalid organization data")
    }

    if err := h.repo.Create(ctx, t); err != nil {
        if stdErrors.Is(err, errors.ErrConflict) {
            return nil, errors.OrganizationAlreadyExists()
        }
        return nil, err
    }

    return t, nil
}
