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

type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error)
	UpdateProfile(ctx context.Context, id domain.AccountID, patch domain.AccountProfilePatch) (*domain.Account, error)
	CreateStorageObject(ctx context.Context, object StorageObject) error
}
