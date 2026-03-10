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

type OrganizationVideoRecord struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ObjectID       uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	CreatedAt      time.Time
	UploadedBy     *uuid.UUID
	SortOrder      int64
}

type LegalDocumentAnalysisLease struct {
	JobID          uuid.UUID
	DocumentID     uuid.UUID
	OrganizationID uuid.UUID
	ObjectID       uuid.UUID
	Bucket         string
	ObjectKey      string
	FileName       string
	MimeType       *string
	Provider       string
	Attempts       int
}

type OrganizationRepository interface {
	Create(ctx context.Context, t *domain.Organization) error
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
	UpdateProfile(ctx context.Context, id domain.OrganizationID, patch domain.OrganizationProfilePatch) (*domain.Organization, error)
	CreateStorageObject(ctx context.Context, object StorageObject) error
	CreateOrganizationVideo(ctx context.Context, organizationID uuid.UUID, objectID uuid.UUID, uploadedBy *uuid.UUID, createdAt time.Time) (*OrganizationVideoRecord, error)
	ListOrganizationVideos(ctx context.Context, organizationID uuid.UUID) ([]OrganizationVideoRecord, error)
	ListOrganizationVideoObjectIDs(ctx context.Context, organizationID uuid.UUID) ([]uuid.UUID, error)
	GetCooperationApplication(ctx context.Context, organizationID domain.OrganizationID) (*domain.CooperationApplication, error)
	SaveCooperationApplication(ctx context.Context, application *domain.CooperationApplication) (*domain.CooperationApplication, error)
	CreateOrganizationLegalDocument(ctx context.Context, document *domain.OrganizationLegalDocument) (*domain.OrganizationLegalDocument, error)
	GetOrganizationLegalDocumentByID(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocument, error)
	ListOrganizationLegalDocuments(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationLegalDocument, error)
	GetOrganizationLegalDocumentAnalysis(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocumentAnalysis, error)
	EnsureOrganizationLegalDocumentAnalysis(ctx context.Context, document *domain.OrganizationLegalDocument, provider string, now time.Time) error
	LeaseNextOrganizationLegalDocumentAnalysisJob(ctx context.Context, now time.Time, leaseFor time.Duration) (*LegalDocumentAnalysisLease, error)
	CompleteOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider string, result LegalDocumentAnalysisResult, completedAt time.Time) error
	FailOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider, errMessage string, retryAt time.Time) error
}
