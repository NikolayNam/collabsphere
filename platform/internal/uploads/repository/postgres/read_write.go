package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type uploadRow struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	OrganizationID     *uuid.UUID `gorm:"column:organization_id"`
	ObjectID           uuid.UUID  `gorm:"column:object_id"`
	CreatedByAccountID uuid.UUID  `gorm:"column:created_by_account_id"`
	Purpose            string     `gorm:"column:purpose"`
	Status             string     `gorm:"column:status"`
	Bucket             string     `gorm:"column:bucket"`
	ObjectKey          string     `gorm:"column:object_key"`
	FileName           string     `gorm:"column:file_name"`
	ContentType        *string    `gorm:"column:content_type"`
	DeclaredSizeBytes  int64      `gorm:"column:declared_size_bytes"`
	ActualSizeBytes    *int64     `gorm:"column:actual_size_bytes"`
	ChecksumSHA256     *string    `gorm:"column:checksum_sha256"`
	Metadata           string     `gorm:"column:metadata"`
	ErrorCode          *string    `gorm:"column:error_code"`
	ErrorMessage       *string    `gorm:"column:error_message"`
	ResultKind         *string    `gorm:"column:result_kind"`
	ResultID           *uuid.UUID `gorm:"column:result_id"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
	ExpiresAt          *time.Time `gorm:"column:expires_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

func (r *Repo) Create(ctx context.Context, upload *uploaddomain.Upload) error {
	if upload == nil {
		return fault.Validation("Upload is required")
	}
	payload := map[string]any{
		"id":                    upload.ID,
		"organization_id":       upload.OrganizationID,
		"object_id":             upload.ObjectID,
		"created_by_account_id": upload.CreatedByAccountID,
		"purpose":               string(upload.Purpose),
		"status":                string(upload.Status),
		"bucket":                upload.Bucket,
		"object_key":            upload.ObjectKey,
		"file_name":             upload.FileName,
		"content_type":          upload.ContentType,
		"declared_size_bytes":   upload.DeclaredSizeBytes,
		"actual_size_bytes":     upload.ActualSizeBytes,
		"checksum_sha256":       upload.ChecksumSHA256,
		"metadata":              jsonbExpr(upload.Metadata),
		"error_code":            upload.ErrorCode,
		"error_message":         upload.ErrorMessage,
		"result_kind":           nullableResultKind(upload.ResultKind),
		"result_id":             upload.ResultID,
		"completed_at":          upload.CompletedAt,
		"expires_at":            upload.ExpiresAt,
		"created_at":            upload.CreatedAt,
		"updated_at":            upload.UpdatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("storage.uploads").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return fault.Conflict("Upload session already exists")
		}
		if isForeignKeyViolation(err) {
			return fault.Validation("Upload session references invalid entities")
		}
		return err
	}
	return nil
}

func (r *Repo) GetByID(ctx context.Context, uploadID uuid.UUID) (*uploaddomain.Upload, error) {
	return r.getOne(ctx, "id = ?", uploadID)
}

func (r *Repo) GetByObjectID(ctx context.Context, objectID uuid.UUID) (*uploaddomain.Upload, error) {
	return r.getOne(ctx, "object_id = ?", objectID)
}

func (r *Repo) MarkReady(ctx context.Context, uploadID uuid.UUID, actualSizeBytes *int64, resultKind uploaddomain.ResultKind, resultID uuid.UUID, completedAt time.Time, updatedAt time.Time) (*uploaddomain.Upload, error) {
	result := r.dbFrom(ctx).WithContext(ctx).
		Table("storage.uploads").
		Where("id = ?", uploadID).
		Updates(map[string]any{
			"status":            string(uploaddomain.StatusReady),
			"actual_size_bytes": actualSizeBytes,
			"error_code":        nil,
			"error_message":     nil,
			"result_kind":       string(resultKind),
			"result_id":         resultID,
			"completed_at":      completedAt,
			"updated_at":        updatedAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, uploadID)
}

func (r *Repo) MarkFailed(ctx context.Context, uploadID uuid.UUID, errorCode, errorMessage string, updatedAt time.Time) (*uploaddomain.Upload, error) {
	result := r.dbFrom(ctx).WithContext(ctx).
		Table("storage.uploads").
		Where("id = ?", uploadID).
		Updates(map[string]any{
			"status":        string(uploaddomain.StatusFailed),
			"error_code":    nullableTrimmedString(errorCode),
			"error_message": nullableTrimmedString(errorMessage),
			"updated_at":    updatedAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, uploadID)
}

func (r *Repo) getOne(ctx context.Context, where string, value any) (*uploaddomain.Upload, error) {
	var row uploadRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("storage.uploads").
		Select("id, organization_id, object_id, created_by_account_id, purpose, status, bucket, object_key, file_name, content_type, declared_size_bytes, actual_size_bytes, checksum_sha256, metadata::text AS metadata, error_code, error_message, result_kind, result_id, completed_at, expires_at, created_at, updated_at").
		Where(where, value).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainUpload(row)
}

func toDomainUpload(row uploadRow) (*uploaddomain.Upload, error) {
	metadata := map[string]any{}
	if row.Metadata != "" {
		if err := json.Unmarshal([]byte(row.Metadata), &metadata); err != nil {
			return nil, err
		}
	}
	var resultKind *uploaddomain.ResultKind
	if row.ResultKind != nil && *row.ResultKind != "" {
		value := uploaddomain.ResultKind(*row.ResultKind)
		resultKind = &value
	}
	return &uploaddomain.Upload{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		ObjectID:           row.ObjectID,
		CreatedByAccountID: row.CreatedByAccountID,
		Purpose:            uploaddomain.Purpose(row.Purpose),
		Status:             uploaddomain.Status(row.Status),
		Bucket:             row.Bucket,
		ObjectKey:          row.ObjectKey,
		FileName:           row.FileName,
		ContentType:        row.ContentType,
		DeclaredSizeBytes:  row.DeclaredSizeBytes,
		ActualSizeBytes:    row.ActualSizeBytes,
		ChecksumSHA256:     row.ChecksumSHA256,
		Metadata:           metadata,
		ErrorCode:          row.ErrorCode,
		ErrorMessage:       row.ErrorMessage,
		ResultKind:         resultKind,
		ResultID:           row.ResultID,
		CompletedAt:        row.CompletedAt,
		ExpiresAt:          row.ExpiresAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}

func jsonbExpr(value map[string]any) any {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte("{}")
	}
	return gorm.Expr("?::jsonb", string(data))
}

func nullableResultKind(value *uploaddomain.ResultKind) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

func nullableTrimmedString(value string) any {
	trimmed := value
	if trimmed == "" {
		return nil
	}
	return trimmed
}
