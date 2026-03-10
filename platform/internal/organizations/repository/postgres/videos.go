package postgres

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/google/uuid"
)

func (r *OrganizationRepo) CreateOrganizationVideo(ctx context.Context, organizationID uuid.UUID, objectID uuid.UUID, uploadedBy *uuid.UUID, createdAt time.Time) (*appports.OrganizationVideoRecord, error) {
	payload := map[string]any{
		"id":              uuid.New(),
		"organization_id": organizationID,
		"object_id":       objectID,
		"sort_order":      0,
		"created_at":      createdAt,
		"uploaded_by":     uploadedBy,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("org.organization_videos").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, apperrors.InvalidInput("Organization video object does not exist or does not belong to organization")
		}
		return nil, err
	}
	items, err := r.ListOrganizationVideos(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ObjectID == objectID {
			return &item, nil
		}
	}
	return nil, nil
}

func (r *OrganizationRepo) ListOrganizationVideos(ctx context.Context, organizationID uuid.UUID) ([]appports.OrganizationVideoRecord, error) {
	var rows []struct {
		ID             uuid.UUID  `gorm:"column:id"`
		OrganizationID uuid.UUID  `gorm:"column:organization_id"`
		ObjectID       uuid.UUID  `gorm:"column:object_id"`
		FileName       string     `gorm:"column:file_name"`
		ContentType    *string    `gorm:"column:content_type"`
		SizeBytes      int64      `gorm:"column:size_bytes"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UploadedBy     *uuid.UUID `gorm:"column:uploaded_by"`
		SortOrder      int64      `gorm:"column:sort_order"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_videos AS ov").
		Select("ov.id, ov.organization_id, ov.object_id, so.file_name, so.content_type, so.size_bytes, ov.created_at, ov.uploaded_by, ov.sort_order").
		Joins("JOIN storage.objects AS so ON so.id = ov.object_id").
		Where("ov.organization_id = ? AND ov.deleted_at IS NULL AND so.deleted_at IS NULL", organizationID).
		Order("ov.sort_order ASC, ov.created_at ASC, ov.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]appports.OrganizationVideoRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, appports.OrganizationVideoRecord{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			ObjectID:       row.ObjectID,
			FileName:       row.FileName,
			ContentType:    row.ContentType,
			SizeBytes:      row.SizeBytes,
			CreatedAt:      row.CreatedAt,
			UploadedBy:     row.UploadedBy,
			SortOrder:      row.SortOrder,
		})
	}
	return out, nil
}

func (r *OrganizationRepo) ListOrganizationVideoObjectIDs(ctx context.Context, organizationID uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.ListOrganizationVideos(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ObjectID)
	}
	return out, nil
}
