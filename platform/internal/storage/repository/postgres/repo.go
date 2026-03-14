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

const maxListedFilesRows = 500

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

func (r *Repo) GetAccountAvatarObjectID(ctx context.Context, accountID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "iam.accounts", "avatar_object_id", "id = ? AND deleted_at IS NULL", accountID)
}

func (r *Repo) GetAccountVideoObjectID(ctx context.Context, accountID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "iam.account_videos", "object_id", "account_id = ? AND id = ? AND deleted_at IS NULL", accountID, videoID)
}

func (r *Repo) GetOrganizationLogoObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "org.organizations", "logo_object_id", "id = ? AND deleted_at IS NULL", organizationID)
}

func (r *Repo) GetOrganizationVideoObjectID(ctx context.Context, organizationID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "org.organization_videos", "object_id", "organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID, videoID)
}

func (r *Repo) GetCooperationPriceListObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "org.cooperation_applications", "price_list_object_id", "organization_id = ?", organizationID)
}

func (r *Repo) GetOrganizationLegalDocumentObjectID(ctx context.Context, organizationID, documentID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "org.organization_legal_documents", "object_id", "organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID, documentID)
}

func (r *Repo) GetProductImportSourceObjectID(ctx context.Context, organizationID, batchID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "catalog.product_import_batches", "source_object_id", "organization_id = ? AND id = ?", organizationID, batchID)
}

func (r *Repo) GetProductVideoObjectID(ctx context.Context, organizationID, productID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "catalog.product_videos", "object_id", "organization_id = ? AND product_id = ? AND id = ? AND deleted_at IS NULL", organizationID, productID, videoID)
}

func (r *Repo) GetConferenceChannelID(ctx context.Context, conferenceID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "collab.conferences", "channel_id", "id = ?", conferenceID)
}

func (r *Repo) GetConferenceRecordingObjectID(ctx context.Context, conferenceID, recordingID uuid.UUID) (*uuid.UUID, error) {
	return r.queryOptionalObjectID(ctx, "collab.conference_recordings", "object_id", "conference_id = ? AND id = ?", conferenceID, recordingID)
}

func (r *Repo) ListConferenceRecordings(ctx context.Context, conferenceID uuid.UUID) ([]storageapp.ConferenceRecordingFile, error) {
	var rows []struct {
		RecordingID  uuid.UUID  `gorm:"column:recording_id"`
		ConferenceID uuid.UUID  `gorm:"column:conference_id"`
		ObjectID     uuid.UUID  `gorm:"column:object_id"`
		FileName     string     `gorm:"column:file_name"`
		ContentType  *string    `gorm:"column:content_type"`
		SizeBytes    int64      `gorm:"column:size_bytes"`
		CreatedAt    time.Time  `gorm:"column:created_at"`
		CreatedBy    *uuid.UUID `gorm:"column:created_by"`
		DurationSec  *int32     `gorm:"column:duration_sec"`
		MimeType     *string    `gorm:"column:mime_type"`
	}
	if err := r.db.WithContext(ctx).
		Table("collab.conference_recordings AS cr").
		Select("cr.id AS recording_id, cr.conference_id, cr.object_id, so.file_name, so.content_type, so.size_bytes, cr.created_at, cr.created_by, cr.duration_sec, cr.mime_type").
		Joins("JOIN storage.objects AS so ON so.id = cr.object_id").
		Where("cr.conference_id = ? AND so.deleted_at IS NULL", conferenceID).
		Order("cr.created_at DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]storageapp.ConferenceRecordingFile, 0, len(rows))
	for _, row := range rows {
		out = append(out, storageapp.ConferenceRecordingFile{
			RecordingID:  row.RecordingID,
			ConferenceID: row.ConferenceID,
			ObjectID:     row.ObjectID,
			FileName:     row.FileName,
			ContentType:  row.ContentType,
			SizeBytes:    row.SizeBytes,
			CreatedAt:    row.CreatedAt,
			CreatedBy:    row.CreatedBy,
			DurationSec:  row.DurationSec,
			MimeType:     row.MimeType,
		})
	}
	return out, nil
}

func (r *Repo) ChannelHasAttachmentObject(ctx context.Context, channelID, objectID uuid.UUID) (bool, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM collab.message_attachments AS ma
		JOIN collab.messages AS m ON m.id = ma.message_id
		WHERE m.channel_id = @channel_id
		  AND ma.object_id = @object_id
		  AND m.deleted_at IS NULL
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("channel_id", channelID), sql.Named("object_id", objectID)).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repo) AccountOwnsAvatar(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("iam.accounts").
		Where("id = ? AND avatar_object_id = ? AND deleted_at IS NULL", accountID, objectID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repo) AccountOwnsVideo(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("iam.account_videos").
		Where("account_id = ? AND object_id = ? AND deleted_at IS NULL", accountID, objectID).
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
		SELECT organization_id FROM org.organization_videos
		WHERE object_id = @object_id AND deleted_at IS NULL
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
		SELECT organization_id FROM catalog.product_videos
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

func (r *Repo) AccountHasAnyOrganizationAccess(ctx context.Context, accountID uuid.UUID, organizationIDs []uuid.UUID) (bool, error) {
	if accountID == uuid.Nil || len(organizationIDs) == 0 {
		return false, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Table("iam.memberships").
		Where("account_id = ? AND organization_id IN ? AND is_active = true AND deleted_at IS NULL", accountID, organizationIDs).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
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
		UNION ALL
		SELECT so.id AS object_id,
		       so.organization_id,
		       so.file_name,
		       so.content_type,
		       so.size_bytes,
		       so.created_at,
		       'account_video' AS source_type,
		       av.id AS source_id
		FROM iam.account_videos AS av
		JOIN storage.objects AS so ON so.id = av.object_id
		WHERE av.account_id = @account_id AND av.deleted_at IS NULL AND so.deleted_at IS NULL
		ORDER BY created_at DESC
	`
	if err := r.db.WithContext(ctx).Raw(query, sql.Named("account_id", accountID)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return trimListedFiles(mapListedFiles(rows), maxListedFilesRows), nil
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
		       'organization_video' AS source_type,
		       ov.id AS source_id
		FROM org.organization_videos AS ov
		JOIN storage.objects AS so ON so.id = ov.object_id
		WHERE ov.organization_id = @organization_id AND ov.deleted_at IS NULL AND so.deleted_at IS NULL
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
		       'product_video' AS source_type,
		       pv.id AS source_id
		FROM catalog.product_videos AS pv
		JOIN storage.objects AS so ON so.id = pv.object_id
		WHERE pv.organization_id = @organization_id AND pv.deleted_at IS NULL AND so.deleted_at IS NULL
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
	return trimListedFiles(mapListedFiles(rows), maxListedFilesRows), nil
}

func (r *Repo) queryOptionalObjectID(ctx context.Context, table, column, where string, args ...any) (*uuid.UUID, error) {
	var row struct {
		ObjectID *uuid.UUID `gorm:"column:object_id"`
	}
	if err := r.db.WithContext(ctx).
		Table(table).
		Select(column+" AS object_id").
		Where(where, args...).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return row.ObjectID, nil
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

func trimListedFiles(values []storageapp.ListedFile, limit int) []storageapp.ListedFile {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}
