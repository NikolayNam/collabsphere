package postgres

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
)

type uploadQueueRow struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	OrganizationID     *uuid.UUID `gorm:"column:organization_id"`
	CreatedByAccountID uuid.UUID  `gorm:"column:created_by_account_id"`
	Purpose            string     `gorm:"column:purpose"`
	Status             string     `gorm:"column:status"`
	FileName           string     `gorm:"column:file_name"`
	ContentType        *string    `gorm:"column:content_type"`
	DeclaredSizeBytes  int64      `gorm:"column:declared_size_bytes"`
	ErrorCode          *string    `gorm:"column:error_code"`
	ErrorMessage       *string    `gorm:"column:error_message"`
	ResultKind         *string    `gorm:"column:result_kind"`
	ResultID           *uuid.UUID `gorm:"column:result_id"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

func (r *Repo) ListUploadQueue(ctx context.Context, query domain.UploadQueueQuery) ([]domain.UploadQueueItem, int, error) {
	db := r.dbFrom(ctx).WithContext(ctx).Table("storage.uploads")
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Purpose != nil {
		db = db.Where("purpose = ?", *query.Purpose)
	}
	if query.OrganizationID != nil {
		db = db.Where("organization_id = ?", *query.OrganizationID)
	}
	if query.CreatedByAccountID != nil {
		db = db.Where("created_by_account_id = ?", *query.CreatedByAccountID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]uploadQueueRow, 0, query.Limit)
	if err := db.
		Select("id, organization_id, created_by_account_id, purpose, status, file_name, content_type, declared_size_bytes, error_code, error_message, result_kind, result_id, created_at, updated_at").
		Order("created_at DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]domain.UploadQueueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UploadQueueItem{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			CreatedByAccountID: row.CreatedByAccountID,
			Purpose:            row.Purpose,
			Status:             row.Status,
			FileName:           row.FileName,
			ContentType:        row.ContentType,
			DeclaredSizeBytes:  row.DeclaredSizeBytes,
			ErrorCode:          row.ErrorCode,
			ErrorMessage:       row.ErrorMessage,
			ResultKind:         row.ResultKind,
			ResultID:           row.ResultID,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		})
	}

	return items, int(total), nil
}
