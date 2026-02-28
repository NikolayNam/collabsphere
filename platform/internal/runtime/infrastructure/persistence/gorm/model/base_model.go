package model

import (
	"time"

	"github.com/google/uuid"
)

type UUIDPK struct {
	ID uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
}

type Timestamps struct {
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamptz;null;autoUpdateTime:false"`
}

type Blame struct {
	CreatedBy *uuid.UUID `gorm:"column:created_by;type:uuid;index"`
	UpdatedBy *uuid.UUID `gorm:"column:updated_by;type:uuid;index"`
}
