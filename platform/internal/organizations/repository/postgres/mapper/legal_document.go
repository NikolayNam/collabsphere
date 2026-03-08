package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDomainOrganizationLegalDocument(m *dbmodel.OrganizationLegalDocument) (*domain.OrganizationLegalDocument, error) {
	if m == nil {
		return nil, nil
	}
	organizationID, err := domain.OrganizationIDFromUUID(m.OrganizationID)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateOrganizationLegalDocument(domain.RehydrateOrganizationLegalDocumentParams{
		ID:                  m.ID,
		OrganizationID:      organizationID,
		DocumentType:        m.DocumentType,
		Status:              m.Status,
		ObjectID:            m.ObjectID,
		Title:               m.Title,
		UploadedByAccountID: m.UploadedByAccountID,
		ReviewerAccountID:   m.ReviewerAccountID,
		ReviewNote:          m.ReviewNote,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		ReviewedAt:          m.ReviewedAt,
		DeletedAt:           m.DeletedAt,
	})
}

func ToDBOrganizationLegalDocument(d *domain.OrganizationLegalDocument) *dbmodel.OrganizationLegalDocument {
	if d == nil {
		return nil
	}
	return &dbmodel.OrganizationLegalDocument{
		ID:                  d.ID(),
		OrganizationID:      d.OrganizationID().UUID(),
		DocumentType:        d.DocumentType(),
		Status:              string(d.Status()),
		ObjectID:            d.ObjectID(),
		Title:               d.Title(),
		UploadedByAccountID: d.UploadedByAccountID(),
		ReviewerAccountID:   d.ReviewerAccountID(),
		ReviewNote:          d.ReviewNote(),
		CreatedAt:           d.CreatedAt(),
		UpdatedAt:           d.UpdatedAt(),
		ReviewedAt:          d.ReviewedAt(),
		DeletedAt:           d.DeletedAt(),
	}
}
