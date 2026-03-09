package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	storageapp "github.com/NikolayNam/collabsphere/internal/storage/application"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

type listedFileRow struct {
	ObjectID       uuid.UUID  `gorm:"column:object_id"`
	OrganizationID *uuid.UUID `gorm:"column:organization_id"`
	FileName       string     `gorm:"column:file_name"`
	ContentType    *string    `gorm:"column:content_type"`
	SizeBytes      int64      `gorm:"column:size_bytes"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	SourceType     string     `gorm:"column:source_type"`
	SourceID       *uuid.UUID `gorm:"column:source_id"`
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) GetObjectByID(ctx context.Context, objectID uuid.UUID) (*storageapp.StoredObject, error) {
	var row struct {
		ID             uuid.UUID  `gorm:"column:id"`
		Bucket         string     `gorm:"column:bucket"`
		ObjectKey      string     `gorm:"column:object_key"`
		FileName       string     `gorm:"column:file_name"`
		ContentType    *string    `gorm:"column:content_type"`
		SizeBytes      int64      `gorm:"column:size_bytes"`
		OrganizationID *uuid.UUID `gorm:"column:organization_id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
	}
	if err := r.db.WithContext(ctx).
		Table("storage.objects").
		Select("id, bucket, object_key, file_name, content_type, size_bytes, organization_id, created_at").
		Where("id = ? AND deleted_at IS NULL", objectID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &storageapp.StoredObject{
		ID:             row.ID,
		Bucket:         row.Bucket,
		ObjectKey:      row.ObjectKey,
		FileName:       row.FileName,
		ContentType:    row.ContentType,
		SizeBytes:      row.SizeBytes,
		OrganizationID: row.OrganizationID,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repo) AccountOwnsAvatar(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("iam.accounts").
		Where("id = ? AND avatar_object_id = ? AND deleted_at IS NULL", accountID, objectID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repo) ListRelatedOrganizationIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error) {
	var rows []struct {
		OrganizationID uuid.UUID `gorm:"column:organization_id"`
	}
	query := `
		SELECT id AS organization_id FROM org.organizations
		WHERE logo_object_id = @object_id AND deleted_at IS NULL
		UNION
		SELECT organization_id FROM org.cooperation_applications
		WHERE price_list_object_id = @object_id
		UNION
		SELECT organization_id FROM org.organization_legal_documents
		WHERE object_id = @object_id AND deleted_at IS NULL
		UNION
		SELECT organization_id FROM catalog.product_import_batches
		WHERE source_object_id = @object_id
		UNION
		SELECT organization_id FROM catalog.product_images
		WHERE object_id = @object_id AND deleted_at IS NULL
		UNION
		SELECT organization_id FROM sales.order_documents
		WHERE object_id = @object_id AND deleted_at IS NULL
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("object_id", objectID)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.OrganizationID)
	}
	return out, nil
}

func (r *Repo) ListRelatedChannelIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error) {
	var rows []struct {
		ChannelID uuid.UUID `gorm:"column:channel_id"`
	}
	query := `
		SELECT m.channel_id AS channel_id
		FROM collab.message_attachments AS ma
		JOIN collab.messages AS m ON m.id = ma.message_id
		WHERE ma.object_id = @object_id AND m.deleted_at IS NULL
		UNION
		SELECT c.channel_id AS channel_id
		FROM collab.conference_recordings AS cr
		JOIN collab.conferences AS c ON c.id = cr.conference_id
		WHERE cr.object_id = @object_id
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("object_id", objectID)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ChannelID)
	}
	return out, nil
}

func (r *Repo) ListAccountFiles(ctx context.Context, accountID uuid.UUID) ([]storageapp.ListedFile, error) {
	var rows []listedFileRow
	query := `
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'account_avatar' AS source_type,
		       a.id AS source_id
		FROM iam.accounts AS a
		JOIN storage.objects AS so ON so.id = a.avatar_object_id
		WHERE a.id = @account_id AND a.deleted_at IS NULL AND so.deleted_at IS NULL
		ORDER BY created_at DESC
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("account_id", accountID)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return mapListedFiles(rows), nil
}

func (r *Repo) ListOrganizationFiles(ctx context.Context, organizationID uuid.UUID) ([]storageapp.ListedFile, error) {
	var rows []listedFileRow
	query := `
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'organization_logo' AS source_type,
		       o.id AS source_id
		FROM org.organizations AS o
		JOIN storage.objects AS so ON so.id = o.logo_object_id
		WHERE o.id = @organization_id AND o.deleted_at IS NULL AND so.deleted_at IS NULL
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'cooperation_price_list' AS source_type,
		       ca.id AS source_id
		FROM org.cooperation_applications AS ca
		JOIN storage.objects AS so ON so.id = ca.price_list_object_id
		WHERE ca.organization_id = @organization_id AND so.deleted_at IS NULL
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'organization_legal_document' AS source_type,
		       ld.id AS source_id
		FROM org.organization_legal_documents AS ld
		JOIN storage.objects AS so ON so.id = ld.object_id
		WHERE ld.organization_id = @organization_id AND ld.deleted_at IS NULL AND so.deleted_at IS NULL
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'product_import_source' AS source_type,
		       pib.id AS source_id
		FROM catalog.product_import_batches AS pib
		JOIN storage.objects AS so ON so.id = pib.source_object_id
		WHERE pib.organization_id = @organization_id AND so.deleted_at IS NULL
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'product_image' AS source_type,
		       pi.id AS source_id
		FROM catalog.product_images AS pi
		JOIN storage.objects AS so ON so.id = pi.object_id
		WHERE pi.organization_id = @organization_id AND pi.deleted_at IS NULL AND so.deleted_at IS NULL
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'order_document' AS source_type,
		       od.id AS source_id
		FROM sales.order_documents AS od
		JOIN storage.objects AS so ON so.id = od.object_id
		WHERE od.organization_id = @organization_id AND od.deleted_at IS NULL AND so.deleted_at IS NULL
		ORDER BY created_at DESC
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("organization_id", organizationID)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return mapListedFiles(rows), nil
}

func mapListedFiles(rows []listedFileRow) []storageapp.ListedFile {
	out := make([]storageapp.ListedFile, 0, len(rows))
	for _, row := range rows {
		out = append(out, storageapp.ListedFile{
			ObjectID:       row.ObjectID,
			OrganizationID: row.OrganizationID,
			FileName:       row.FileName,
			ContentType:    row.ContentType,
			SizeBytes:      row.SizeBytes,
			CreatedAt:      row.CreatedAt,
			SourceType:     row.SourceType,
			SourceID:       row.SourceID,
		})
	}
	return out
}
