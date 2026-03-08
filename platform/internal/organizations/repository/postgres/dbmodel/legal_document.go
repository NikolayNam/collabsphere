package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationLegalDocument struct {
	ID                  uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID      uuid.UUID  `gorm:"column:organization_id;type:uuid;not null"`
	DocumentType        string     `gorm:"column:document_type;type:varchar(64);not null"`
	Status              string     `gorm:"column:status;type:varchar(32);not null"`
	ObjectID            uuid.UUID  `gorm:"column:object_id;type:uuid;not null"`
	Title               string     `gorm:"column:title;type:varchar(255);not null"`
	UploadedByAccountID *uuid.UUID `gorm:"column:uploaded_by_account_id;type:uuid"`
	ReviewerAccountID   *uuid.UUID `gorm:"column:reviewer_account_id;type:uuid"`
	ReviewNote          *string    `gorm:"column:review_note;type:text"`
	CreatedAt           time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt           *time.Time `gorm:"column:updated_at;type:timestamptz"`
	ReviewedAt          *time.Time `gorm:"column:reviewed_at;type:timestamptz"`
	DeletedAt           *time.Time `gorm:"column:deleted_at;type:timestamptz"`
}

func (OrganizationLegalDocument) TableName() string { return "org.organization_legal_documents" }
