package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type channelRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	GroupID        uuid.UUID  `gorm:"column:group_id"`
	Slug           string     `gorm:"column:slug"`
	Name           string     `gorm:"column:name"`
	Description    *string    `gorm:"column:description"`
	IsDefault      bool       `gorm:"column:is_default"`
	LastMessageSeq int64      `gorm:"column:last_message_seq"`
	CreatedBy      *uuid.UUID `gorm:"column:created_by"`
	UpdatedBy      *uuid.UUID `gorm:"column:updated_by"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
}

func (r *Repo) ProvisionDefaultChannel(ctx context.Context, groupID, createdBy uuid.UUID, now time.Time) error {
	if groupID == uuid.Nil {
		return fmt.Errorf("group id is required")
	}
	name := "General"
	slug := "general"
	_, err := r.CreateChannel(ctx, collabdomain.Channel{
		ID:        uuid.New(),
		GroupID:   groupID,
		Slug:      slug,
		Name:      name,
		IsDefault: true,
		CreatedBy: uuidPtr(createdBy),
		CreatedAt: now,
	}, nil, nil, nil)
	if err != nil && isUniqueViolation(err) {
		return nil
	}
	return err
}

func (r *Repo) CreateChannel(ctx context.Context, channel collabdomain.Channel, adminAccountIDs []uuid.UUID, organizationIDs, accountIDs []uuid.UUID) (*collabdomain.Channel, error) {
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}
	if channel.CreatedAt.IsZero() {
		channel.CreatedAt = time.Now().UTC()
	}
	updatedAt := channel.CreatedAt
	row := map[string]any{
		"id":               channel.ID,
		"group_id":         channel.GroupID,
		"slug":             strings.TrimSpace(channel.Slug),
		"name":             strings.TrimSpace(channel.Name),
		"description":      channel.Description,
		"is_default":       channel.IsDefault,
		"last_message_seq": channel.LastMessageSeq,
		"created_at":       channel.CreatedAt,
		"updated_at":       &updatedAt,
		"created_by":       channel.CreatedBy,
		"updated_by":       channel.UpdatedBy,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.channels").Create(row).Error; err != nil {
		return nil, err
	}
	for _, adminID := range uniqueUUIDs(adminAccountIDs) {
		if adminID == uuid.Nil {
			continue
		}
		if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.channel_admins").Create(map[string]any{
			"channel_id": channel.ID,
			"account_id": adminID,
			"created_at": channel.CreatedAt,
			"created_by": channel.CreatedBy,
		}).Error; err != nil {
			return nil, err
		}
	}
	for _, orgID := range uniqueUUIDs(organizationIDs) {
		if orgID == uuid.Nil {
			continue
		}
		if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.channel_organizations").Create(map[string]any{
			"channel_id":      channel.ID,
			"organization_id": orgID,
			"created_at":     channel.CreatedAt,
		}).Error; err != nil {
			return nil, err
		}
	}
	for _, accID := range uniqueUUIDs(accountIDs) {
		if accID == uuid.Nil {
			continue
		}
		if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.channel_accounts").Create(map[string]any{
			"channel_id": channel.ID,
			"account_id": accID,
			"created_at": channel.CreatedAt,
		}).Error; err != nil {
			return nil, err
		}
	}
	return r.GetChannelByID(ctx, channel.ID)
}

func (r *Repo) GetChannelByID(ctx context.Context, channelID uuid.UUID) (*collabdomain.Channel, error) {
	var row channelRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channels").
		Select("id, group_id, slug, name, description, is_default, last_message_seq, created_by, updated_by, created_at, updated_at").
		Where("id = ? AND deleted_at IS NULL", channelID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapChannel(row), nil
}

func (r *Repo) ListAllChannelIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channels").
		Select("id").
		Where("deleted_at IS NULL").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *Repo) ListChannelsByGroup(ctx context.Context, groupID uuid.UUID) ([]collabdomain.Channel, error) {
	return r.listChannelsByGroup(ctx, groupID, nil)
}

func (r *Repo) ListChannelsByGroupForAccount(ctx context.Context, groupID, accountID uuid.UUID) ([]collabdomain.Channel, error) {
	return r.listChannelsByGroup(ctx, groupID, &accountID)
}

func (r *Repo) listChannelsByGroup(ctx context.Context, groupID uuid.UUID, accountID *uuid.UUID) ([]collabdomain.Channel, error) {
	var rows []channelRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.channels").
		Select("id, group_id, slug, name, description, is_default, last_message_seq, created_by, updated_by, created_at, updated_at").
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("is_default DESC, created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if accountID == nil || *accountID == uuid.Nil {
		out := make([]collabdomain.Channel, 0, len(rows))
		for _, row := range rows {
			out = append(out, *mapChannel(row))
		}
		return out, nil
	}
	access, err := r.ResolveGroupAccessForAccount(ctx, groupID, *accountID)
	if err != nil || !access.Allowed {
		return nil, err
	}
	out := make([]collabdomain.Channel, 0, len(rows))
	for _, row := range rows {
		hasOrgs, hasAccounts, err := r.channelHasVisibilityRestrictions(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		if !hasOrgs && !hasAccounts {
			out = append(out, *mapChannel(row))
			continue
		}
		allowed, err := r.accountPassesChannelVisibility(ctx, row.ID, *accountID, access.OrganizationIDs)
		if err != nil {
			return nil, err
		}
		if allowed {
			out = append(out, *mapChannel(row))
		}
	}
	return out, nil
}

func (r *Repo) CreateStorageObject(ctx context.Context, object collabdomain.StorageObject) error {
	if object.ID == uuid.Nil {
		object.ID = uuid.New()
	}
	if object.CreatedAt.IsZero() {
		object.CreatedAt = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("storage.objects").Create(map[string]any{
		"id":              object.ID,
		"organization_id": object.OrganizationID,
		"bucket":          object.Bucket,
		"object_key":      object.ObjectKey,
		"file_name":       object.FileName,
		"content_type":    object.ContentType,
		"size_bytes":      object.SizeBytes,
		"checksum_sha256": object.ChecksumSHA256,
		"created_at":      object.CreatedAt,
	}).Error
}

func (r *Repo) GetStorageObject(ctx context.Context, objectID uuid.UUID) (*collabdomain.StorageObject, error) {
	var row struct {
		ID             uuid.UUID  `gorm:"column:id"`
		OrganizationID *uuid.UUID `gorm:"column:organization_id"`
		Bucket         string     `gorm:"column:bucket"`
		ObjectKey      string     `gorm:"column:object_key"`
		FileName       string     `gorm:"column:file_name"`
		ContentType    *string    `gorm:"column:content_type"`
		SizeBytes      int64      `gorm:"column:size_bytes"`
		ChecksumSHA256 *string    `gorm:"column:checksum_sha256"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("storage.objects").
		Select("id, organization_id, bucket, object_key, file_name, content_type, size_bytes, checksum_sha256, created_at").
		Where("id = ? AND deleted_at IS NULL", objectID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &collabdomain.StorageObject{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Bucket:         row.Bucket,
		ObjectKey:      row.ObjectKey,
		FileName:       row.FileName,
		ContentType:    row.ContentType,
		SizeBytes:      row.SizeBytes,
		ChecksumSHA256: row.ChecksumSHA256,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func mapChannel(row channelRow) *collabdomain.Channel {
	return &collabdomain.Channel{
		ID:             row.ID,
		GroupID:        row.GroupID,
		Slug:           row.Slug,
		Name:           row.Name,
		Description:    row.Description,
		IsDefault:      row.IsDefault,
		LastMessageSeq: row.LastMessageSeq,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func (r *Repo) ValidateChannelVisibilityInGroup(ctx context.Context, groupID uuid.UUID, organizationIDs, accountIDs []uuid.UUID) error {
	for _, orgID := range organizationIDs {
		if orgID == uuid.Nil {
			continue
		}
		var count int64
		if err := r.dbFrom(ctx).WithContext(ctx).
			Table("iam.group_organization_members").
			Where("group_id = ? AND organization_id = ? AND deleted_at IS NULL AND is_active = TRUE", groupID, orgID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("organization %s is not a member of the group", orgID)
		}
	}
	for _, accID := range accountIDs {
		if accID == uuid.Nil {
			continue
		}
		var count int64
		if err := r.dbFrom(ctx).WithContext(ctx).
			Table("iam.group_account_members").
			Where("group_id = ? AND account_id = ? AND deleted_at IS NULL AND is_active = TRUE", groupID, accID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("account %s is not a member of the group", accID)
		}
	}
	return nil
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	v := id
	return &v
}
