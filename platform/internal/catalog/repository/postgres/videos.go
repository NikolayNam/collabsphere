package postgres

import (
	"context"
	"fmt"
	"time"

	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	"github.com/google/uuid"
)

func (r *CatalogRepo) CreateProductVideo(ctx context.Context, organizationID, productID, objectID uuid.UUID, uploadedBy *uuid.UUID, createdAt time.Time) (*ports.ProductVideoRecord, error) {
	payload := map[string]any{
		"id":              uuid.New(),
		"organization_id": organizationID,
		"product_id":      productID,
		"object_id":       objectID,
		"sort_order":      0,
		"created_at":      createdAt,
		"uploaded_by":     uploadedBy,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("catalog.product_videos").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", catalogerrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, catalogerrors.InvalidInput("Product video object or product does not exist")
		}
		return nil, err
	}
	items, err := r.ListProductVideos(ctx, organizationID, productID)
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

func (r *CatalogRepo) ListProductVideos(ctx context.Context, organizationID, productID uuid.UUID) ([]ports.ProductVideoRecord, error) {
	var rows []struct {
		ID             uuid.UUID  `gorm:"column:id"`
		OrganizationID uuid.UUID  `gorm:"column:organization_id"`
		ProductID      uuid.UUID  `gorm:"column:product_id"`
		ObjectID       uuid.UUID  `gorm:"column:object_id"`
		FileName       string     `gorm:"column:file_name"`
		ContentType    *string    `gorm:"column:content_type"`
		SizeBytes      int64      `gorm:"column:size_bytes"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UploadedBy     *uuid.UUID `gorm:"column:uploaded_by"`
		SortOrder      int64      `gorm:"column:sort_order"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_videos AS pv").
		Select("pv.id, pv.organization_id, pv.product_id, pv.object_id, so.file_name, so.content_type, so.size_bytes, pv.created_at, pv.uploaded_by, pv.sort_order").
		Joins("JOIN storage.objects AS so ON so.id = pv.object_id").
		Where("pv.organization_id = ? AND pv.product_id = ? AND pv.deleted_at IS NULL AND so.deleted_at IS NULL", organizationID, productID).
		Order("pv.sort_order ASC, pv.created_at ASC, pv.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ports.ProductVideoRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, ports.ProductVideoRecord{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			ProductID:      row.ProductID,
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

func (r *CatalogRepo) ListProductVideoObjectIDs(ctx context.Context, organizationID, productID uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.ListProductVideos(ctx, organizationID, productID)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ObjectID)
	}
	return out, nil
}

func (r *CatalogRepo) ListProductVideoObjectIDsByProduct(ctx context.Context, organizationID uuid.UUID, productIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
		ObjectID  uuid.UUID `gorm:"column:object_id"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_videos AS pv").
		Select("pv.product_id, pv.object_id").
		Joins("JOIN storage.objects AS so ON so.id = pv.object_id").
		Where("pv.organization_id = ? AND pv.product_id IN ? AND pv.deleted_at IS NULL AND so.deleted_at IS NULL", organizationID, productIDs).
		Order("pv.product_id ASC, pv.sort_order ASC, pv.created_at ASC, pv.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ProductID] = append(result[row.ProductID], row.ObjectID)
	}
	return result, nil
}
