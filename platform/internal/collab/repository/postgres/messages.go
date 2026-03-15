package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type messageRow struct {
	ID               uuid.UUID  `gorm:"column:id"`
	ChannelID        uuid.UUID  `gorm:"column:channel_id"`
	ChannelSeq       int64      `gorm:"column:channel_seq"`
	MessageType      string     `gorm:"column:message_type"`
	AuthorType       string     `gorm:"column:author_type"`
	AuthorAccountID  *uuid.UUID `gorm:"column:author_account_id"`
	AuthorGuestID    *uuid.UUID `gorm:"column:author_guest_id"`
	AccountName      *string    `gorm:"column:account_name"`
	GuestName        *string    `gorm:"column:guest_name"`
	Body             string     `gorm:"column:body"`
	ReplyToMessageID *uuid.UUID `gorm:"column:reply_to_message_id"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	EditedAt         *time.Time `gorm:"column:edited_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

type reactionRow struct {
	MessageID uuid.UUID `gorm:"column:message_id"`
	Emoji     string    `gorm:"column:emoji"`
	Count     int64     `gorm:"column:count"`
	Mine      bool      `gorm:"column:mine"`
}

type mentionRow struct {
	MessageID uuid.UUID `gorm:"column:message_id"`
	AccountID uuid.UUID `gorm:"column:account_id"`
}

type attachmentRow struct {
	MessageID      uuid.UUID  `gorm:"column:message_id"`
	ObjectID       uuid.UUID  `gorm:"column:object_id"`
	OrganizationID *uuid.UUID `gorm:"column:organization_id"`
	FileName       string     `gorm:"column:file_name"`
	ContentType    *string    `gorm:"column:content_type"`
	SizeBytes      int64      `gorm:"column:size_bytes"`
	Bucket         string     `gorm:"column:bucket"`
	ObjectKey      string     `gorm:"column:object_key"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
}

func (r *Repo) CreateMessage(ctx context.Context, message collabdomain.Message, mentionIDs, attachmentObjectIDs []uuid.UUID) (*collabdomain.Message, error) {
	if message.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channel id is required")
	}
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().UTC()
	}

	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Transaction(func(tx *gorm.DB) error {
		var seq int64
		update := tx.Raw(`
			UPDATE collab.channels
			SET last_message_seq = last_message_seq + 1,
			    updated_at = ?,
			    updated_by = ?
			WHERE id = ? AND deleted_at IS NULL
			RETURNING last_message_seq
		`, message.CreatedAt, message.AuthorAccountID, message.ChannelID).Scan(&seq)
		if update.Error != nil {
			return update.Error
		}
		if seq == 0 {
			return gorm.ErrRecordNotFound
		}
		message.ChannelSeq = seq

		if err := tx.Table("collab.messages").Create(map[string]any{
			"id":                  message.ID,
			"channel_id":          message.ChannelID,
			"channel_seq":         message.ChannelSeq,
			"message_type":        string(message.Type),
			"author_type":         string(message.AuthorType),
			"author_account_id":   message.AuthorAccountID,
			"author_guest_id":     message.AuthorGuestID,
			"body":                message.Body,
			"reply_to_message_id": message.ReplyToMessageID,
			"created_at":          message.CreatedAt,
			"edited_at":           message.EditedAt,
			"deleted_at":          message.DeletedAt,
		}).Error; err != nil {
			return err
		}

		for _, mentionID := range uniqueUUIDs(mentionIDs) {
			if err := tx.Table("collab.message_mentions").Create(map[string]any{
				"message_id": message.ID,
				"account_id": mentionID,
				"created_at": message.CreatedAt,
			}).Error; err != nil {
				return err
			}
		}

		for _, objectID := range uniqueUUIDs(attachmentObjectIDs) {
			object, err := r.getStorageObjectWithDB(tx, ctx, objectID)
			if err != nil {
				return err
			}
			if object == nil {
				return gorm.ErrRecordNotFound
			}
			if err := tx.Table("collab.message_attachments").Create(map[string]any{
				"message_id":      message.ID,
				"object_id":       object.ID,
				"organization_id": object.OrganizationID,
				"created_at":      message.CreatedAt,
				"created_by":      message.AuthorAccountID,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return r.GetMessageByID(ctx, message.ChannelID, message.ID, authdomain.AnonymousPrincipal())
}

func (r *Repo) GetMessageByID(ctx context.Context, channelID, messageID uuid.UUID, principal authdomain.Principal) (*collabdomain.Message, error) {
	messages, err := r.listMessages(ctx, channelID, principal, func(db *gorm.DB) *gorm.DB {
		return db.Where("m.id = ?", messageID)
	})
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil
	}
	return &messages[0], nil
}

func (r *Repo) ListMessages(ctx context.Context, channelID uuid.UUID, principal authdomain.Principal, visibleFromSeq int64, limit int) ([]collabdomain.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return r.listMessages(ctx, channelID, principal, func(db *gorm.DB) *gorm.DB {
		query := db.Where("m.channel_id = ?", channelID)
		if visibleFromSeq > 0 {
			query = query.Where("m.channel_seq >= ?", visibleFromSeq)
		}
		return query.Order("m.channel_seq ASC").Limit(limit)
	})
}

// ListRecentMessagesForChannel returns the last N messages for a channel without access control, oldest first.
// Used by Redis broker init to preload channel history.
func (r *Repo) ListRecentMessagesForChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]collabdomain.Message, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	msgs, err := r.listMessages(ctx, channelID, authdomain.AnonymousPrincipal(), func(db *gorm.DB) *gorm.DB {
		return db.Where("m.channel_id = ? AND m.deleted_at IS NULL", channelID).
			Order("m.channel_seq DESC").
			Limit(limit)
	})
	if err != nil || len(msgs) == 0 {
		return msgs, err
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *Repo) listMessages(ctx context.Context, channelID uuid.UUID, principal authdomain.Principal, scope func(*gorm.DB) *gorm.DB) ([]collabdomain.Message, error) {
	base := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.messages AS m").
		Select("m.id, m.channel_id, m.channel_seq, m.message_type, m.author_type, m.author_account_id, m.author_guest_id, a.display_name AS account_name, g.display_name AS guest_name, m.body, m.reply_to_message_id, m.created_at, m.edited_at, m.deleted_at").
		Joins("LEFT JOIN iam.accounts AS a ON a.id = m.author_account_id").
		Joins("LEFT JOIN auth.guest_identities AS g ON g.id = m.author_guest_id")
	if scope != nil {
		base = scope(base)
	}

	var rows []messageRow
	if err := base.Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	messageIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		messageIDs = append(messageIDs, row.ID)
	}

	mentionsByMessage, err := r.loadMentions(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	attachmentsByMessage, err := r.loadAttachments(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	reactionsByMessage, err := r.loadReactions(ctx, messageIDs, principal)
	if err != nil {
		return nil, err
	}

	out := make([]collabdomain.Message, 0, len(rows))
	for _, row := range rows {
		authorName := row.AccountName
		if authorName == nil {
			authorName = row.GuestName
		}
		out = append(out, collabdomain.Message{
			ID:               row.ID,
			ChannelID:        row.ChannelID,
			ChannelSeq:       row.ChannelSeq,
			Type:             collabdomain.MessageType(row.MessageType),
			AuthorType:       collabdomain.ActorType(row.AuthorType),
			AuthorAccountID:  row.AuthorAccountID,
			AuthorGuestID:    row.AuthorGuestID,
			AuthorName:       authorName,
			Body:             row.Body,
			ReplyToMessageID: row.ReplyToMessageID,
			CreatedAt:        row.CreatedAt,
			EditedAt:         row.EditedAt,
			DeletedAt:        row.DeletedAt,
			Mentions:         mentionsByMessage[row.ID],
			Attachments:      attachmentsByMessage[row.ID],
			Reactions:        reactionsByMessage[row.ID],
		})
	}
	return out, nil
}

func (r *Repo) UpdateMessageBody(ctx context.Context, messageID uuid.UUID, body string, editedBy *uuid.UUID, mentionIDs []uuid.UUID, editedAt time.Time, principal authdomain.Principal) (*collabdomain.Message, error) {
	if editedAt.IsZero() {
		editedAt = time.Now().UTC()
	}
	var channelID uuid.UUID
	if err := r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current struct {
			ChannelID uuid.UUID `gorm:"column:channel_id"`
			Body      string    `gorm:"column:body"`
		}
		if err := tx.Table("collab.messages").
			Select("channel_id, body").
			Where("id = ?", messageID).
			Take(&current).Error; err != nil {
			return err
		}
		channelID = current.ChannelID

		if err := tx.Table("collab.message_revisions").Create(map[string]any{
			"id":         uuid.New(),
			"message_id": messageID,
			"body":       current.Body,
			"edited_by":  editedBy,
			"created_at": editedAt,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("collab.messages").Where("id = ?", messageID).Updates(map[string]any{
			"body":      body,
			"edited_at": editedAt,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("collab.message_mentions").Where("message_id = ?", messageID).Delete(nil).Error; err != nil {
			return err
		}
		for _, mentionID := range uniqueUUIDs(mentionIDs) {
			if err := tx.Table("collab.message_mentions").Create(map[string]any{
				"message_id": messageID,
				"account_id": mentionID,
				"created_at": editedAt,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return r.GetMessageByID(ctx, channelID, messageID, principal)
}

func (r *Repo) DeleteMessage(ctx context.Context, messageID uuid.UUID, deletedAt time.Time) error {
	if deletedAt.IsZero() {
		deletedAt = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("collab.messages").Where("id = ? AND deleted_at IS NULL", messageID).Updates(map[string]any{
		"deleted_at": deletedAt,
		"edited_at":  deletedAt,
	}).Error
}

func (r *Repo) UpsertReadCursor(ctx context.Context, principal authdomain.Principal, channelID uuid.UUID, lastReadSeq int64, at time.Time) error {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	values := map[string]any{
		"channel_id":    channelID,
		"actor_type":    string(principal.SubjectType),
		"account_id":    nil,
		"guest_id":      nil,
		"last_read_seq": lastReadSeq,
		"last_read_at":  at,
	}
	switch {
	case principal.IsAccount():
		values["account_id"] = principal.AccountID
	case principal.IsGuest():
		values["guest_id"] = principal.GuestID
	default:
		return fmt.Errorf("principal is not authenticated")
	}
	return r.dbFrom(ctx).WithContext(ctx).Exec(`
		INSERT INTO collab.channel_read_cursors (channel_id, actor_type, account_id, guest_id, last_read_seq, last_read_at)
		VALUES (@channel_id, @actor_type, @account_id, @guest_id, @last_read_seq, @last_read_at)
		ON CONFLICT ON CONSTRAINT uq_collab_channel_read_cursors_actor
		DO UPDATE SET last_read_seq = EXCLUDED.last_read_seq, last_read_at = EXCLUDED.last_read_at
	`, values).Error
}

func (r *Repo) ToggleReaction(ctx context.Context, principal authdomain.Principal, messageID uuid.UUID, emoji string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	actorType := string(principal.SubjectType)
	var accountID any
	var guestID any
	switch {
	case principal.IsAccount():
		accountID = principal.AccountID
	case principal.IsGuest():
		guestID = principal.GuestID
	default:
		return fmt.Errorf("principal is not authenticated")
	}

	db := r.dbFrom(ctx).WithContext(ctx)
	result := db.Table("collab.message_reactions").Where("message_id = ? AND actor_type = ? AND account_id IS NOT DISTINCT FROM ? AND guest_id IS NOT DISTINCT FROM ? AND emoji = ?", messageID, actorType, accountID, guestID, emoji).Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return db.Table("collab.message_reactions").Create(map[string]any{
		"message_id": messageID,
		"actor_type": actorType,
		"account_id": accountID,
		"guest_id":   guestID,
		"emoji":      emoji,
		"created_at": now,
	}).Error
}

func (r *Repo) loadMentions(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	var rows []mentionRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.message_mentions").
		Select("message_id, account_id").
		Where("message_id IN ?", messageIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID][]uuid.UUID, len(messageIDs))
	for _, row := range rows {
		out[row.MessageID] = append(out[row.MessageID], row.AccountID)
	}
	return out, nil
}

func (r *Repo) loadAttachments(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]collabdomain.Attachment, error) {
	var rows []attachmentRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.message_attachments AS ma").
		Select("ma.message_id, so.id AS object_id, so.organization_id, so.file_name, so.content_type, so.size_bytes, so.bucket, so.object_key, ma.created_at").
		Joins("JOIN storage.objects AS so ON so.id = ma.object_id").
		Where("ma.message_id IN ?", messageIDs).
		Order("ma.created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID][]collabdomain.Attachment, len(messageIDs))
	for _, row := range rows {
		out[row.MessageID] = append(out[row.MessageID], collabdomain.Attachment{
			ObjectID:       row.ObjectID,
			OrganizationID: row.OrganizationID,
			FileName:       row.FileName,
			ContentType:    row.ContentType,
			SizeBytes:      row.SizeBytes,
			Bucket:         row.Bucket,
			ObjectKey:      row.ObjectKey,
			CreatedAt:      row.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repo) loadReactions(ctx context.Context, messageIDs []uuid.UUID, principal authdomain.Principal) (map[uuid.UUID][]collabdomain.ReactionSummary, error) {
	query := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.message_reactions AS mr").
		Select("mr.message_id, mr.emoji, COUNT(*) AS count, FALSE AS mine").
		Where("mr.message_id IN ?", messageIDs).
		Group("mr.message_id, mr.emoji")

	if principal.IsAccount() {
		query = r.dbFrom(ctx).WithContext(ctx).
			Table("collab.message_reactions AS mr").
			Select("mr.message_id, mr.emoji, COUNT(*) AS count, BOOL_OR(mr.actor_type = 'account' AND mr.account_id = ?) AS mine", principal.AccountID).
			Where("mr.message_id IN ?", messageIDs).
			Group("mr.message_id, mr.emoji")
	} else if principal.IsGuest() {
		query = r.dbFrom(ctx).WithContext(ctx).
			Table("collab.message_reactions AS mr").
			Select("mr.message_id, mr.emoji, COUNT(*) AS count, BOOL_OR(mr.actor_type = 'guest' AND mr.guest_id = ?) AS mine", principal.GuestID).
			Where("mr.message_id IN ?", messageIDs).
			Group("mr.message_id, mr.emoji")
	}

	var rows []reactionRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID][]collabdomain.ReactionSummary, len(messageIDs))
	for _, row := range rows {
		out[row.MessageID] = append(out[row.MessageID], collabdomain.ReactionSummary{
			Emoji: row.Emoji,
			Count: row.Count,
			Mine:  row.Mine,
		})
	}
	for id := range out {
		sort.Slice(out[id], func(i, j int) bool {
			return out[id][i].Emoji < out[id][j].Emoji
		})
	}
	return out, nil
}

func (r *Repo) GetAccountChatAttachmentTotalSize(ctx context.Context, accountID uuid.UUID) (int64, error) {
	if accountID == uuid.Nil {
		return 0, nil
	}
	var total int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.message_attachments AS ma").
		Select("COALESCE(SUM(so.size_bytes), 0)").
		Joins("JOIN storage.objects AS so ON so.id = ma.object_id AND so.deleted_at IS NULL").
		Where("ma.created_by = ?", accountID).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repo) GetOrganizationChatAttachmentTotalSize(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	if organizationID == uuid.Nil {
		return 0, nil
	}
	var total int64
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.message_attachments AS ma").
		Select("COALESCE(SUM(so.size_bytes), 0)").
		Joins("JOIN storage.objects AS so ON so.id = ma.object_id AND so.deleted_at IS NULL").
		Where("ma.organization_id = ?", organizationID).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

type attachmentLimitsRow struct {
	DocumentLimitBytes int64 `gorm:"column:document_limit_bytes"`
	PhotoLimitBytes    int64 `gorm:"column:photo_limit_bytes"`
	VideoLimitBytes    int64 `gorm:"column:video_limit_bytes"`
	TotalLimitBytes    int64 `gorm:"column:total_limit_bytes"`
}

func (r *Repo) GetEffectiveAttachmentLimits(ctx context.Context, accountID uuid.UUID, organizationID *uuid.UUID) (document, photo, video, total int64, err error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	// Resolution order: account > organization > platform
	if accountID != uuid.Nil {
		var row attachmentLimitsRow
		if err := db.Table("storage.attachment_limits").
			Select("document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes").
			Where("scope_type = ? AND scope_id = ?", "account", accountID).
			Take(&row).Error; err == nil {
			return row.DocumentLimitBytes, row.PhotoLimitBytes, row.VideoLimitBytes, row.TotalLimitBytes, nil
		}
	}
	if organizationID != nil && *organizationID != uuid.Nil {
		var row attachmentLimitsRow
		if err := db.Table("storage.attachment_limits").
			Select("document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes").
			Where("scope_type = ? AND scope_id = ?", "organization", *organizationID).
			Take(&row).Error; err == nil {
			return row.DocumentLimitBytes, row.PhotoLimitBytes, row.VideoLimitBytes, row.TotalLimitBytes, nil
		}
	}
	var row attachmentLimitsRow
	if err := db.Table("storage.attachment_limits").
		Select("document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes").
		Where("scope_type = ? AND scope_id IS NULL", "platform").
		Take(&row).Error; err != nil {
		return 0, 0, 0, 0, err
	}
	return row.DocumentLimitBytes, row.PhotoLimitBytes, row.VideoLimitBytes, row.TotalLimitBytes, nil
}

func (r *Repo) getStorageObjectWithDB(db *gorm.DB, ctx context.Context, objectID uuid.UUID) (*collabdomain.StorageObject, error) {
	if db == nil {
		db = r.dbFrom(ctx)
	}
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
	if err := db.WithContext(ctx).Table("storage.objects").
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
