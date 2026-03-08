package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDomainCooperationApplication(m *dbmodel.CooperationApplication) (*domain.CooperationApplication, error) {
	if m == nil {
		return nil, nil
	}
	organizationID, err := domain.OrganizationIDFromUUID(m.OrganizationID)
	if err != nil {
		return nil, err
	}
	salesChannels, err := domain.UnmarshalSalesChannels(m.SalesChannels)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateCooperationApplication(domain.RehydrateCooperationApplicationParams{
		ID:                    m.ID,
		OrganizationID:        organizationID,
		Status:                m.Status,
		ConfirmationEmail:     m.ConfirmationEmail,
		CompanyName:           m.CompanyName,
		RepresentedCategories: m.RepresentedCategories,
		MinimumOrderAmount:    m.MinimumOrderAmount,
		DeliveryGeography:     m.DeliveryGeography,
		SalesChannels:         salesChannels,
		StorefrontURL:         m.StorefrontURL,
		ContactFirstName:      m.ContactFirstName,
		ContactLastName:       m.ContactLastName,
		ContactJobTitle:       m.ContactJobTitle,
		PriceListObjectID:     m.PriceListObjectID,
		ContactEmail:          m.ContactEmail,
		ContactPhone:          m.ContactPhone,
		PartnerCode:           m.PartnerCode,
		ReviewNote:            m.ReviewNote,
		ReviewerAccountID:     m.ReviewerAccountID,
		SubmittedAt:           m.SubmittedAt,
		ReviewedAt:            m.ReviewedAt,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	})
}

func ToDBCooperationApplication(a *domain.CooperationApplication) (*dbmodel.CooperationApplication, error) {
	if a == nil {
		return nil, nil
	}
	salesChannels, err := domain.MarshalSalesChannels(a.SalesChannels())
	if err != nil {
		return nil, err
	}
	return &dbmodel.CooperationApplication{
		ID:                    a.ID(),
		OrganizationID:        a.OrganizationID().UUID(),
		Status:                string(a.Status()),
		ConfirmationEmail:     a.ConfirmationEmail(),
		CompanyName:           a.CompanyName(),
		RepresentedCategories: a.RepresentedCategories(),
		MinimumOrderAmount:    a.MinimumOrderAmount(),
		DeliveryGeography:     a.DeliveryGeography(),
		SalesChannels:         salesChannels,
		StorefrontURL:         a.StorefrontURL(),
		ContactFirstName:      a.ContactFirstName(),
		ContactLastName:       a.ContactLastName(),
		ContactJobTitle:       a.ContactJobTitle(),
		PriceListObjectID:     a.PriceListObjectID(),
		ContactEmail:          a.ContactEmail(),
		ContactPhone:          a.ContactPhone(),
		PartnerCode:           a.PartnerCode(),
		ReviewNote:            a.ReviewNote(),
		ReviewerAccountID:     a.ReviewerAccountID(),
		SubmittedAt:           a.SubmittedAt(),
		ReviewedAt:            a.ReviewedAt(),
		CreatedAt:             a.CreatedAt(),
		UpdatedAt:             a.UpdatedAt(),
	}, nil
}
