package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	CreateStorageObject(ctx context.Context, obj *StorageObject) error
	GetStorageObjectByID(ctx context.Context, organizationID uuid.UUID, objectID uuid.UUID) (*StorageObject, error)

	CreatePaymentImport(ctx context.Context, imp *PaymentImport) error
	UpdatePaymentImport(ctx context.Context, imp *PaymentImport) error
	GetPaymentImportByID(ctx context.Context, organizationID, importID uuid.UUID) (*PaymentImport, error)
	GetPaymentImportBySourceObjectID(ctx context.Context, organizationID, sourceObjectID uuid.UUID) (*PaymentImport, error)

	AddPaymentImportItems(ctx context.Context, importID uuid.UUID, items []PaymentImportItem) error
	ListPaymentImportItems(ctx context.Context, importID uuid.UUID) ([]PaymentImportItem, error)

	ListPaymentAccounts(ctx context.Context, organizationID uuid.UUID) ([]PaymentAccount, error)
}

type PaymentAccount struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Kind           string
	CurrencyCode   string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
