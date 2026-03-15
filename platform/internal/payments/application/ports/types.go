package ports

import (
	"time"

	"github.com/google/uuid"
)

// PaymentImportDirection — направление платежей в файле
type PaymentImportDirection string

const (
	DirectionIncoming PaymentImportDirection = "incoming"
	DirectionOutgoing PaymentImportDirection = "outgoing"
	DirectionMixed    PaymentImportDirection = "mixed"
)

// PaymentImportStatus — статус импорта
type PaymentImportStatus string

const (
	ImportStatusPending    PaymentImportStatus = "pending"
	ImportStatusParsing    PaymentImportStatus = "parsing"
	ImportStatusParsed    PaymentImportStatus = "parsed"
	ImportStatusMapping   PaymentImportStatus = "mapping"
	ImportStatusApplying  PaymentImportStatus = "applying"
	ImportStatusCompleted PaymentImportStatus = "completed"
	ImportStatusFailed    PaymentImportStatus = "failed"
	ImportStatusCancelled PaymentImportStatus = "cancelled"
)

// ParsedPaymentItem — нормализованная строка платежа после парсинга
type ParsedPaymentItem struct {
	RowIndex     int
	Direction    string // incoming | outgoing
	AmountCents  int64
	CurrencyCode string
	OccurredAt   time.Time
	Counterparty *string
	Purpose      *string
	RawData      map[string]any
}

// PaymentImportAnalysis — результат анализа файла
type PaymentImportAnalysis struct {
	FormatCode string
	Direction  PaymentImportDirection
	Items      []ParsedPaymentItem
}

// StorageObject — объект в хранилище (для чтения файла)
type StorageObject struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Bucket         string
	ObjectKey      string
	FileName       string
	ContentType    *string
	SizeBytes      int64
}

// PaymentImport — сессия импорта
type PaymentImport struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	SourceObjectID     uuid.UUID
	CreatedByAccountID uuid.UUID
	FormatCode         string
	Direction          *string
	Status             PaymentImportStatus
	TotalItems         *int
	AppliedItems       int
	ErrorItems         int
	AnalysisResult     map[string]any
	StartedAt          time.Time
	FinishedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}

// PaymentImportItem — строка импорта
type PaymentImportItem struct {
	ID               uuid.UUID
	ImportID         uuid.UUID
	RowIndex         int
	Direction        string
	AmountCents      int64
	CurrencyCode     string
	OccurredAt       time.Time
	Counterparty     *string
	Purpose          *string
	RawData          map[string]any
	MappedAccountID  *uuid.UUID
	TransactionID    *uuid.UUID
	ErrorCode        *string
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}
