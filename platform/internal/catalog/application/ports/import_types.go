package ports

import (
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type StorageObject struct {
	ID             uuid.UUID
	OrganizationID orgdomain.OrganizationID
	Bucket         string
	ObjectKey      string
	FileName       string
	ContentType    *string
	SizeBytes      int64
	ChecksumSHA256 *string
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

type ProductImportStatus string

const (
	ProductImportStatusPending    ProductImportStatus = "pending"
	ProductImportStatusProcessing ProductImportStatus = "processing"
	ProductImportStatusCompleted  ProductImportStatus = "completed"
	ProductImportStatusFailed     ProductImportStatus = "failed"
	ProductImportStatusCancelled  ProductImportStatus = "cancelled"
)

type ProductImportBatch struct {
	ID                 uuid.UUID
	OrganizationID     orgdomain.OrganizationID
	SourceObjectID     uuid.UUID
	CreatedByAccountID accdomain.AccountID
	Status             ProductImportStatus
	TotalRows          *int
	ProcessedRows      int
	SuccessRows        int
	ErrorRows          int
	StartedBy          *string
	StartedAt          time.Time
	FinishedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	Mode               *string
	ResultSummary      map[string]any
}

type ProductImportErrorRecord struct {
	ID        uuid.UUID
	BatchID   uuid.UUID
	RowNo     *int
	Code      *string
	Message   string
	Details   map[string]any
	CreatedAt time.Time
}
