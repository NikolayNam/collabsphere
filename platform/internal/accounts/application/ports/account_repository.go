package ports

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
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

type AccountVideoRecord struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	ObjectID    uuid.UUID
	FileName    string
	ContentType *string
	SizeBytes   int64
	CreatedAt   time.Time
	SortOrder   int64
}

type AccountKYCProfileRecord struct {
	AccountID        uuid.UUID
	Status           string
	LegalName        *string
	CountryCode      *string
	DocumentNumber   *string
	ResidenceAddress *string
	ReviewNote       *string
	ReviewerAccount  *uuid.UUID
	SubmittedAt      *time.Time
	ReviewedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type AccountKYCDocumentRecord struct {
	ID              uuid.UUID
	AccountID       uuid.UUID
	ObjectID        uuid.UUID
	DocumentType    string
	Title           string
	Status          string
	ReviewNote      *string
	ReviewerAccount *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	ReviewedAt      *time.Time
}

type AccountKYCProfilePatch struct {
	Status           string
	LegalName        *string
	CountryCode      *string
	DocumentNumber   *string
	ResidenceAddress *string
	ReviewNote       *string
	ReviewerAccount  *uuid.UUID
	SubmittedAt      *time.Time
	ReviewedAt       *time.Time
	UpdatedAt        time.Time
}

type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error)
	UpdateProfile(ctx context.Context, id domain.AccountID, patch domain.AccountProfilePatch) (*domain.Account, error)
	CreateStorageObject(ctx context.Context, object StorageObject) error
	CreateAccountVideo(ctx context.Context, accountID uuid.UUID, objectID uuid.UUID, createdAt time.Time) (*AccountVideoRecord, error)
	ListAccountVideos(ctx context.Context, accountID uuid.UUID) ([]AccountVideoRecord, error)
	ListAccountVideoObjectIDs(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error)
	GetAccountKYCProfile(ctx context.Context, accountID uuid.UUID) (*AccountKYCProfileRecord, error)
	UpsertAccountKYCProfile(ctx context.Context, accountID uuid.UUID, patch AccountKYCProfilePatch) (*AccountKYCProfileRecord, error)
	ListAccountKYCDocuments(ctx context.Context, accountID uuid.UUID) ([]AccountKYCDocumentRecord, error)
	GetAccountKYCDocumentByObjectID(ctx context.Context, accountID, objectID uuid.UUID) (*AccountKYCDocumentRecord, error)
	CreateAccountKYCDocument(ctx context.Context, item AccountKYCDocumentRecord) (*AccountKYCDocumentRecord, error)
}
