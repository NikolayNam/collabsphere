package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NikolayNam/collabsphere/internal/platformops/application/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type attachmentLimitRow struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	ScopeType          string     `gorm:"column:scope_type"`
	ScopeID            *uuid.UUID `gorm:"column:scope_id"`
	DocumentLimitBytes int64      `gorm:"column:document_limit_bytes"`
	PhotoLimitBytes    int64      `gorm:"column:photo_limit_bytes"`
	VideoLimitBytes    int64      `gorm:"column:video_limit_bytes"`
	TotalLimitBytes    int64      `gorm:"column:total_limit_bytes"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (r *Repo) List(ctx context.Context, scopeType *string, scopeID *uuid.UUID) ([]ports.AttachmentLimit, error) {
	db := r.dbFrom(ctx).WithContext(ctx).Table("storage.attachment_limits")
	if scopeType != nil {
		db = db.Where("scope_type = ?", *scopeType)
	}
	if scopeID != nil {
		db = db.Where("scope_id = ?", *scopeID)
	}
	var rows []attachmentLimitRow
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ports.AttachmentLimit, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToLimit(row))
	}
	return out, nil
}

func (r *Repo) GetPlatform(ctx context.Context) (*ports.AttachmentLimit, error) {
	var row attachmentLimitRow
	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Table("storage.attachment_limits").
		Where("scope_type = ? AND scope_id IS NULL", "platform").
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	limit := rowToLimit(row)
	return &limit, nil
}

func (r *Repo) GetByScope(ctx context.Context, scopeType string, scopeID uuid.UUID) (*ports.AttachmentLimit, error) {
	var row attachmentLimitRow
	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Table("storage.attachment_limits").
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	limit := rowToLimit(row)
	return &limit, nil
}

func (r *Repo) UpsertPlatform(ctx context.Context, limit ports.AttachmentLimit, now time.Time) (*ports.AttachmentLimit, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	var row attachmentLimitRow
	err := db.Table("storage.attachment_limits").
		Where("scope_type = ? AND scope_id IS NULL", "platform").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = limitToRow(limit)
			row.ID = uuid.New()
			row.ScopeType = "platform"
			row.ScopeID = nil
			row.CreatedAt = now
			row.UpdatedAt = now
			if err := db.Table("storage.attachment_limits").Create(&row).Error; err != nil {
				return nil, err
			}
			out := rowToLimit(row)
			return &out, nil
		}
		return nil, err
	}
	row.DocumentLimitBytes = limit.DocumentLimitBytes
	row.PhotoLimitBytes = limit.PhotoLimitBytes
	row.VideoLimitBytes = limit.VideoLimitBytes
	row.TotalLimitBytes = limit.TotalLimitBytes
	row.UpdatedAt = now
	if err := db.Table("storage.attachment_limits").Save(&row).Error; err != nil {
		return nil, err
	}
	out := rowToLimit(row)
	return &out, nil
}

func (r *Repo) UpsertByScope(ctx context.Context, scopeType string, scopeID uuid.UUID, limit ports.AttachmentLimit, now time.Time) (*ports.AttachmentLimit, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	var row attachmentLimitRow
	err := db.Table("storage.attachment_limits").
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = limitToRow(limit)
			row.ID = uuid.New()
			row.ScopeType = scopeType
			row.ScopeID = &scopeID
			row.CreatedAt = now
			row.UpdatedAt = now
			if err := db.Table("storage.attachment_limits").Create(&row).Error; err != nil {
				return nil, err
			}
			out := rowToLimit(row)
			return &out, nil
		}
		return nil, err
	}
	row.DocumentLimitBytes = limit.DocumentLimitBytes
	row.PhotoLimitBytes = limit.PhotoLimitBytes
	row.VideoLimitBytes = limit.VideoLimitBytes
	row.TotalLimitBytes = limit.TotalLimitBytes
	row.UpdatedAt = now
	if err := db.Table("storage.attachment_limits").Save(&row).Error; err != nil {
		return nil, err
	}
	out := rowToLimit(row)
	return &out, nil
}

func (r *Repo) DeleteByScope(ctx context.Context, scopeType string, scopeID uuid.UUID) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	return db.Table("storage.attachment_limits").
		Where("scope_type = ? AND scope_id = ?", scopeType, scopeID).
		Delete(nil).Error
}

func rowToLimit(row attachmentLimitRow) ports.AttachmentLimit {
	return ports.AttachmentLimit{
		ID:                 row.ID,
		ScopeType:          row.ScopeType,
		ScopeID:            row.ScopeID,
		DocumentLimitBytes: row.DocumentLimitBytes,
		PhotoLimitBytes:    row.PhotoLimitBytes,
		VideoLimitBytes:    row.VideoLimitBytes,
		TotalLimitBytes:    row.TotalLimitBytes,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func limitToRow(limit ports.AttachmentLimit) attachmentLimitRow {
	return attachmentLimitRow{
		ID:                 limit.ID,
		ScopeType:          limit.ScopeType,
		ScopeID:            limit.ScopeID,
		DocumentLimitBytes: limit.DocumentLimitBytes,
		PhotoLimitBytes:    limit.PhotoLimitBytes,
		VideoLimitBytes:    limit.VideoLimitBytes,
		TotalLimitBytes:    limit.TotalLimitBytes,
		CreatedAt:          limit.CreatedAt,
		UpdatedAt:          limit.UpdatedAt,
	}
}
