package create_with_owner

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgErrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	orgPorts "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/shared/tx"
)

type Handler struct {
	tx         tx.Manager
	orgRepo    orgPorts.OrganizationRepository
	memRepo    memberPorts.MembershipRepository
	categories orgPorts.ProductCategoryProvisioner
}

func New(txm tx.Manager, orgRepo orgPorts.OrganizationRepository, memRepo memberPorts.MembershipRepository, categories orgPorts.ProductCategoryProvisioner) *Handler {
	return &Handler{tx: txm, orgRepo: orgRepo, memRepo: memRepo, categories: categories}
}

func (h *Handler) Handle(ctx context.Context, org *orgDomain.Organization, ownerAccountID accdomain.AccountID, domains []orgDomain.OrganizationDomainDraft) error {
	if org == nil {
		return orgErrors.InvalidInput("Organization is required")
	}
	if ownerAccountID.IsZero() {
		return orgErrors.InvalidInput("Owner account is required")
	}

	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.orgRepo.Create(ctx, org); err != nil {
			return err
		}

		if len(domains) > 0 {
			records, err := orgDomain.BuildOrganizationDomains(org.ID(), domains, nil, org.CreatedAt())
			if err != nil {
				return orgErrors.InvalidInput(err.Error())
			}
			if _, err := h.orgRepo.ReplaceDomains(ctx, org.ID(), records, org.CreatedAt()); err != nil {
				return err
			}
		}

		membership, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
			OrganizationID: org.ID(),
			AccountID:      ownerAccountID,
			Role:           memberDomain.MembershipRoleOwner,
			Now:            org.CreatedAt(),
		})
		if err != nil {
			return orgErrors.InvalidInput("Invalid owner membership")
		}

		if err := h.memRepo.AddMember(ctx, org.ID(), membership); err != nil {
			return err
		}

		if h.categories != nil {
			if err := h.categories.ProvisionDefaults(ctx, org.ID(), org.CreatedAt()); err != nil {
				return err
			}
		}

		return nil
	})
}
