package ports

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type StorageObject struct {
	ID             uuid.UUID
	OrganizationID *uuid.UUID
	Bucket         string
	ObjectKey      string
	FileName       string
	ContentType    *string
	SizeBytes      int64
	ChecksumSHA256 *string
	CreatedAt      time.Time
}

type OrganizationRepository interface {
	Create(ctx context.Context, t *domain.Organization) error
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
	UpdateProfile(ctx context.Context, id domain.OrganizationID, patch domain.OrganizationProfilePatch) (*domain.Organization, error)
	CreateStorageObject(ctx context.Context, object StorageObject) error
	GetCooperationApplication(ctx context.Context, organizationID domain.OrganizationID) (*domain.CooperationApplication, error)
	SaveCooperationApplication(ctx context.Context, application *domain.CooperationApplication) (*domain.CooperationApplication, error)
	CreateOrganizationLegalDocument(ctx context.Context, document *domain.OrganizationLegalDocument) (*domain.OrganizationLegalDocument, error)
	ListOrganizationLegalDocuments(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationLegalDocument, error)
}
