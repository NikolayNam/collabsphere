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

type OrganizationMembershipView struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	LogoObjectID   *uuid.UUID
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	MembershipRole string
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

type OrganizationKYCProfileRecord struct {
	OrganizationID     uuid.UUID
	Status             string
	LegalName          *string
	CountryCode        *string
	RegistrationNumber *string
	TaxID              *string
	ReviewNote         *string
	ReviewerAccountID  *uuid.UUID
	SubmittedAt        *time.Time
	ReviewedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrganizationKYCDocumentRecord struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	ObjectID          uuid.UUID
	DocumentType      string
	Title             string
	Status            string
	ReviewNote        *string
	ReviewerAccountID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	ReviewedAt        *time.Time
}

type OrganizationKYCProfilePatch struct {
	Status             string
	LegalName          *string
	CountryCode        *string
	RegistrationNumber *string
	TaxID              *string
	ReviewNote         *string
	ReviewerAccountID  *uuid.UUID
	SubmittedAt        *time.Time
	ReviewedAt         *time.Time
	UpdatedAt          time.Time
}

type OrganizationRepository interface {
	Create(ctx context.Context, t *domain.Organization) error
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]OrganizationMembershipView, error)
	GetByHostname(ctx context.Context, hostname string) (*domain.Organization, error)
	ListDomains(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationDomain, error)
	ReplaceDomains(ctx context.Context, organizationID domain.OrganizationID, domains []domain.OrganizationDomain, now time.Time) ([]domain.OrganizationDomain, error)
	UpdateProfile(ctx context.Context, id domain.OrganizationID, patch domain.OrganizationProfilePatch) (*domain.Organization, error)
	CreateStorageObject(ctx context.Context, object StorageObject) error
	CreateOrganizationVideo(ctx context.Context, organizationID uuid.UUID, objectID uuid.UUID, uploadedBy *uuid.UUID, createdAt time.Time) (*OrganizationVideoRecord, error)
	ListOrganizationVideos(ctx context.Context, organizationID uuid.UUID) ([]OrganizationVideoRecord, error)
	ListOrganizationVideoObjectIDs(ctx context.Context, organizationID uuid.UUID) ([]uuid.UUID, error)
	GetCooperationApplication(ctx context.Context, organizationID domain.OrganizationID) (*domain.CooperationApplication, error)
	SaveCooperationApplication(ctx context.Context, application *domain.CooperationApplication) (*domain.CooperationApplication, error)
	CreateOrganizationLegalDocument(ctx context.Context, document *domain.OrganizationLegalDocument) (*domain.OrganizationLegalDocument, error)
	GetOrganizationLegalDocumentByID(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocument, error)
	GetOrganizationLegalDocumentByObjectID(ctx context.Context, organizationID domain.OrganizationID, objectID uuid.UUID) (*domain.OrganizationLegalDocument, error)
	ListOrganizationLegalDocuments(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationLegalDocument, error)
	GetOrganizationLegalDocumentAnalysis(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocumentAnalysis, error)
	EnsureOrganizationLegalDocumentAnalysis(ctx context.Context, document *domain.OrganizationLegalDocument, provider string, now time.Time) error
	LeaseNextOrganizationLegalDocumentAnalysisJob(ctx context.Context, now time.Time, leaseFor time.Duration) (*LegalDocumentAnalysisLease, error)
	CompleteOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider string, result LegalDocumentAnalysisResult, completedAt time.Time) error
	FailOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider, errMessage string, retryAt time.Time) error
	GetOrganizationKYCProfile(ctx context.Context, organizationID uuid.UUID) (*OrganizationKYCProfileRecord, error)
	UpsertOrganizationKYCProfile(ctx context.Context, organizationID uuid.UUID, patch OrganizationKYCProfilePatch) (*OrganizationKYCProfileRecord, error)
	ListOrganizationKYCDocuments(ctx context.Context, organizationID uuid.UUID) ([]OrganizationKYCDocumentRecord, error)
	GetOrganizationKYCDocumentByObjectID(ctx context.Context, organizationID, objectID uuid.UUID) (*OrganizationKYCDocumentRecord, error)
	CreateOrganizationKYCDocument(ctx context.Context, item OrganizationKYCDocumentRecord) (*OrganizationKYCDocumentRecord, error)
}
