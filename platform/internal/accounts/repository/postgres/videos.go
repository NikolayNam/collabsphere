package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *AccountRepo) CreateAccountVideo(ctx context.Context, accountID uuid.UUID, objectID uuid.UUID, createdAt time.Time) (*appports.AccountVideoRecord, error) {
	payload := map[string]any{
		"id":         uuid.New(),
		"account_id": accountID,
		"object_id":  objectID,
		"sort_order": 0,
		"created_at": createdAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("iam.account_videos").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, apperrors.InvalidInput("Account video object does not exist")
		}
		return nil, err
	}
	return r.getAccountVideoByObjectID(ctx, accountID, objectID)
}

func (r *AccountRepo) ListAccountVideos(ctx context.Context, accountID uuid.UUID) ([]appports.AccountVideoRecord, error) {
	var rows []struct {
		ID          uuid.UUID  `gorm:"column:id"`
		AccountID   uuid.UUID  `gorm:"column:account_id"`
		ObjectID    uuid.UUID  `gorm:"column:object_id"`
		FileName    string     `gorm:"column:file_name"`
		ContentType *string    `gorm:"column:content_type"`
		SizeBytes   int64      `gorm:"column:size_bytes"`
		CreatedAt   time.Time  `gorm:"column:created_at"`
		SortOrder   int64      `gorm:"column:sort_order"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.account_videos AS av").
		Select("av.id, av.account_id, av.object_id, so.file_name, so.content_type, so.size_bytes, av.created_at, av.sort_order").
		Joins("JOIN storage.objects AS so ON so.id = av.object_id").
		Where("av.account_id = ? AND av.deleted_at IS NULL AND so.deleted_at IS NULL", accountID).
		Order("av.sort_order ASC, av.created_at ASC, av.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]appports.AccountVideoRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, appports.AccountVideoRecord{
			ID:          row.ID,
			AccountID:   row.AccountID,
			ObjectID:    row.ObjectID,
			FileName:    row.FileName,
			ContentType: row.ContentType,
			SizeBytes:   row.SizeBytes,
			CreatedAt:   row.CreatedAt,
			SortOrder:   row.SortOrder,
		})
	}
	return out, nil
}

func (r *AccountRepo) ListAccountVideoObjectIDs(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	items, err := r.ListAccountVideos(ctx, accountID)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ObjectID)
	}
	return out, nil
}

func (r *AccountRepo) getAccountVideoByObjectID(ctx context.Context, accountID uuid.UUID, objectID uuid.UUID) (*appports.AccountVideoRecord, error) {
	items, err := r.ListAccountVideos(ctx, accountID)
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

func (r *AccountRepo) accountVideoExists(ctx context.Context, accountID uuid.UUID, objectID uuid.UUID) (bool, error) {
	var count int64
	if err := r.dbFrom(ctx).WithContext(ctx).Table("iam.account_videos").Where("account_id = ? AND object_id = ? AND deleted_at IS NULL", accountID, objectID).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
