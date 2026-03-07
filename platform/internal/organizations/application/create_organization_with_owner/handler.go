package create_with_owner

import (
	"context"

	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgPorts "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/shared/tx"
)

type Handler struct {
	tx      tx.Manager
	orgRepo orgPorts.OrganizationRepository
	mem     memberPorts.MembershipWriter
}

func New(txm tx.Manager, orgRepo orgPorts.OrganizationRepository, mem memberPorts.MembershipWriter) *Handler {
	return &Handler{tx: txm, orgRepo: orgRepo, mem: mem}
}

func (h *Handler) Handle(ctx context.Context, org *orgDomain.Organization, ownerAccountID string) error {
	// Валидации здесь (org != nil, ownerAccountID, etc.)
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.orgRepo.Create(ctx, org); err != nil {
			return err
		}
		// kind="owner" — можно захардкодить или передать
		if err := h.mem.AddMember(ctx, org.ID(), ownerAccountID, "owner"); err != nil {
			return err
		}
		return nil
	})
}
