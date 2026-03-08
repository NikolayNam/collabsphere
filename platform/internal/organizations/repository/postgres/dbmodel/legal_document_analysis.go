package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationLegalDocumentAnalysis struct {
	ID                   uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	DocumentID           uuid.UUID  `gorm:"column:document_id;type:uuid;not null"`
	OrganizationID       uuid.UUID  `gorm:"column:organization_id;type:uuid;not null"`
	Status               string     `gorm:"column:status;type:varchar(32);not null"`
	Provider             string     `gorm:"column:provider;type:varchar(64);not null"`
	ExtractedText        *string    `gorm:"column:extracted_text;type:text"`
	Summary              *string    `gorm:"column:summary;type:text"`
	ExtractedFieldsJSON  []byte     `gorm:"column:extracted_fields_json;type:jsonb;not null"`
	DetectedDocumentType *string    `gorm:"column:detected_document_type;type:varchar(128)"`
	ConfidenceScore      *float64   `gorm:"column:confidence_score;type:double precision"`
	RequestedAt          time.Time  `gorm:"column:requested_at;type:timestamptz;not null"`
	StartedAt            *time.Time `gorm:"column:started_at;type:timestamptz"`
	CompletedAt          *time.Time `gorm:"column:completed_at;type:timestamptz"`
	UpdatedAt            *time.Time `gorm:"column:updated_at;type:timestamptz"`
	LastError            *string    `gorm:"column:last_error;type:text"`
}

func (OrganizationLegalDocumentAnalysis) TableName() string {
	return "org.organization_legal_document_analysis"
}
