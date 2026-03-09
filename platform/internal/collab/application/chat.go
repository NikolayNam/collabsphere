package application

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

func (s *Service) CreateChannel(ctx context.Context, cmd CreateChannelCmd) (*collabdomain.Channel, error) {
	access, err := s.requireGroupAccountAccess(ctx, cmd.GroupID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	if !access.CanManage {
		return nil, fault.Forbidden("Only group owners can create channels")
	}
	name := strings.TrimSpace(cmd.Name)
	slug := sanitizeSlug(cmd.Slug)
	if name == "" || slug == "" {
		return nil, fault.Validation("Channel name and slug are required")
	}
	now := s.now()
	created, err := s.repo.CreateChannel(ctx, collabdomain.Channel{
		ID:          uuid.New(),
		GroupID:     cmd.GroupID,
		Slug:        slug,
		Name:        name,
		Description: normalizeOptional(cmd.Description),
		CreatedBy:   uuidPtr(cmd.Actor.AccountID),
		UpdatedBy:   uuidPtr(cmd.Actor.AccountID),
		CreatedAt:   now,
	}, cmd.AdminAccountIDs)
	if err != nil {
		if collabpg.IsUnique(err) {
			return nil, fault.Conflict("Channel already exists")
		}
		return nil, fault.Internal("Create channel failed", fault.WithCause(err))
	}
	return created, nil
}

func (s *Service) ListChannels(ctx context.Context, q ListChannelsQuery) ([]collabdomain.Channel, error) {
	if _, err := s.requireGroupAccountAccess(ctx, q.GroupID, q.Actor); err != nil {
		return nil, err
	}
	channels, err := s.repo.ListChannelsByGroup(ctx, q.GroupID)
	if err != nil {
		return nil, fault.Internal("List channels failed", fault.WithCause(err))
	}
	return channels, nil
}

func (s *Service) CreateMessage(ctx context.Context, cmd CreateMessageCmd) (*collabdomain.Message, error) {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	if !access.CanPost {
		return nil, fault.Forbidden("Posting is not allowed")
	}
	if cmd.Actor.IsGuest() && len(cmd.MentionAccountIDs) > 0 {
		return nil, fault.Forbidden("Guests cannot mention accounts")
	}
	body := strings.TrimSpace(cmd.Body)
	attachmentIDs := uniqueUUIDs(cmd.AttachmentObjectIDs)
	if body == "" && len(attachmentIDs) == 0 {
		return nil, fault.Validation("Message body or attachment is required")
	}
	msg := collabdomain.Message{
		ID:               uuid.New(),
		ChannelID:        cmd.ChannelID,
		Type:             collabdomain.MessageTypeUser,
		AuthorType:       actorTypeFromPrincipal(cmd.Actor),
		AuthorAccountID:  principalAccountPtr(cmd.Actor),
		AuthorGuestID:    principalGuestPtr(cmd.Actor),
		Body:             body,
		ReplyToMessageID: cmd.ReplyToMessageID,
		CreatedAt:        s.now(),
	}
	created, err := s.repo.CreateMessage(ctx, msg, uniqueUUIDs(cmd.MentionAccountIDs), attachmentIDs)
	if err != nil {
		return nil, fault.Internal("Create message failed", fault.WithCause(err))
	}
	_ = s.repo.UpsertReadCursor(ctx, cmd.Actor, cmd.ChannelID, created.ChannelSeq, s.now())
	s.publish(ctx, collabdomain.Event{Type: "message.created", ChannelID: cmd.ChannelID, Payload: created})
	return created, nil
}

func (s *Service) ListMessages(ctx context.Context, q ListMessagesQuery) ([]collabdomain.Message, error) {
	access, _, err := s.requireChannelAccess(ctx, q.ChannelID, q.Actor)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.ListMessages(ctx, q.ChannelID, q.Actor, access.VisibleFromSeq, q.Limit)
	if err != nil {
		return nil, fault.Internal("List messages failed", fault.WithCause(err))
	}
	return messages, nil
}

func (s *Service) UpdateMessage(ctx context.Context, cmd UpdateMessageCmd) (*collabdomain.Message, error) {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	current, err := s.repo.GetMessageByID(ctx, cmd.ChannelID, cmd.MessageID, cmd.Actor)
	if err != nil {
		return nil, fault.Internal("Load message failed", fault.WithCause(err))
	}
	if current == nil {
		return nil, fault.NotFound("Message not found")
	}
	if !s.canMutateMessage(access, cmd.Actor, current) {
		return nil, fault.Forbidden("Message update is not allowed")
	}
	if cmd.Actor.IsGuest() && len(cmd.MentionAccountIDs) > 0 {
		return nil, fault.Forbidden("Guests cannot mention accounts")
	}
	body := strings.TrimSpace(cmd.Body)
	if body == "" {
		return nil, fault.Validation("Message body is required")
	}
	updated, err := s.repo.UpdateMessageBody(ctx, cmd.MessageID, body, principalAccountPtr(cmd.Actor), uniqueUUIDs(cmd.MentionAccountIDs), s.now(), cmd.Actor)
	if err != nil {
		return nil, fault.Internal("Update message failed", fault.WithCause(err))
	}
	s.publish(ctx, collabdomain.Event{Type: "message.updated", ChannelID: cmd.ChannelID, Payload: updated})
	return updated, nil
}

func (s *Service) DeleteMessage(ctx context.Context, cmd DeleteMessageCmd) error {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return err
	}
	current, err := s.repo.GetMessageByID(ctx, cmd.ChannelID, cmd.MessageID, cmd.Actor)
	if err != nil {
		return fault.Internal("Load message failed", fault.WithCause(err))
	}
	if current == nil {
		return fault.NotFound("Message not found")
	}
	if !s.canMutateMessage(access, cmd.Actor, current) {
		return fault.Forbidden("Message delete is not allowed")
	}
	if err := s.repo.DeleteMessage(ctx, cmd.MessageID, s.now()); err != nil {
		return fault.Internal("Delete message failed", fault.WithCause(err))
	}
	s.publish(ctx, collabdomain.Event{Type: "message.deleted", ChannelID: cmd.ChannelID, Payload: map[string]any{"messageId": cmd.MessageID}})
	return nil
}

func (s *Service) CreateAttachmentUpload(ctx context.Context, cmd CreateAttachmentUploadCmd) (*CreateAttachmentUploadResult, error) {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	if !access.CanPost {
		return nil, fault.Forbidden("Posting is not allowed")
	}
	if s.storage == nil || strings.TrimSpace(s.storageBucket) == "" {
		return nil, fault.Unavailable("Attachment upload is unavailable")
	}
	fileName := strings.TrimSpace(cmd.FileName)
	if fileName == "" {
		return nil, fault.Validation("fileName is required")
	}
	orgID := cmd.OrganizationID
	if orgID != nil && cmd.Actor.IsGuest() {
		return nil, fault.Forbidden("Guests cannot bind attachment to organization")
	}
	objectID := uuid.New()
	now := s.now()
	objectKey := buildAttachmentObjectKey(cmd.ChannelID, objectID, fileName, now)
	obj := collabdomain.StorageObject{
		ID:             objectID,
		OrganizationID: orgID,
		Bucket:         strings.TrimSpace(s.storageBucket),
		ObjectKey:      objectKey,
		FileName:       sanitizeFileName(fileName),
		ContentType:    normalizeOptional(cmd.ContentType),
		SizeBytes:      cmd.SizeBytes,
		ChecksumSHA256: normalizeOptional(cmd.ChecksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, obj); err != nil {
		return nil, fault.Internal("Create attachment object failed", fault.WithCause(err))
	}
	url, expiresAt, err := s.storage.PresignPutObject(ctx, obj.Bucket, obj.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign attachment upload failed", fault.WithCause(err))
	}
	return &CreateAttachmentUploadResult{ObjectID: obj.ID, UploadURL: url, ExpiresAt: expiresAt, Bucket: obj.Bucket, ObjectKey: obj.ObjectKey, FileName: obj.FileName, SizeBytes: obj.SizeBytes, OrganizationID: obj.OrganizationID, ContentType: obj.ContentType, CreatedAt: obj.CreatedAt}, nil
}

func (s *Service) UploadAttachment(ctx context.Context, cmd UploadAttachmentCmd) (*collabdomain.Attachment, error) {
	if cmd.Body == nil {
		return nil, fault.Validation("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, fault.Validation("file size must be non-negative")
	}
	contentType := cmd.ContentType
	upload, err := s.CreateAttachmentUpload(ctx, CreateAttachmentUploadCmd{
		ChannelID:      cmd.ChannelID,
		Actor:          cmd.Actor,
		OrganizationID: cmd.OrganizationID,
		FileName:       cmd.FileName,
		ContentType:    normalizeOptional(&contentType),
		SizeBytes:      cmd.SizeBytes,
		ChecksumSHA256: nil,
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	if err := s.storage.PutObject(ctx, upload.Bucket, upload.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload attachment failed", fault.WithCause(err))
	}
	attachmentContentType := upload.ContentType
	if attachmentContentType == nil {
		attachmentContentType = &contentType
	}
	return &collabdomain.Attachment{
		ObjectID:       upload.ObjectID,
		OrganizationID: upload.OrganizationID,
		FileName:       upload.FileName,
		ContentType:    attachmentContentType,
		SizeBytes:      upload.SizeBytes,
		Bucket:         upload.Bucket,
		ObjectKey:      upload.ObjectKey,
		CreatedAt:      upload.CreatedAt,
	}, nil
}
func (s *Service) UpdateReadCursor(ctx context.Context, cmd UpdateReadCursorCmd) error {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return err
	}
	if !access.CanRead {
		return fault.Forbidden("Channel access is required")
	}
	if cmd.LastReadSeq < access.VisibleFromSeq {
		cmd.LastReadSeq = access.VisibleFromSeq
	}
	if err := s.repo.UpsertReadCursor(ctx, cmd.Actor, cmd.ChannelID, cmd.LastReadSeq, s.now()); err != nil {
		return fault.Internal("Update read cursor failed", fault.WithCause(err))
	}
	s.publish(ctx, collabdomain.Event{Type: "channel.read_cursor", ChannelID: cmd.ChannelID, Payload: map[string]any{"lastReadSeq": cmd.LastReadSeq}})
	return nil
}

func (s *Service) ToggleReaction(ctx context.Context, cmd ToggleReactionCmd) error {
	access, _, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return err
	}
	if !access.CanRead {
		return fault.Forbidden("Channel access is required")
	}
	emoji := strings.TrimSpace(cmd.Emoji)
	if emoji == "" {
		return fault.Validation("emoji is required")
	}
	if err := s.repo.ToggleReaction(ctx, cmd.Actor, cmd.MessageID, emoji, s.now()); err != nil {
		return fault.Internal("Toggle reaction failed", fault.WithCause(err))
	}
	s.publish(ctx, collabdomain.Event{Type: "message.reaction", ChannelID: cmd.ChannelID, Payload: map[string]any{"messageId": cmd.MessageID, "emoji": emoji}})
	return nil
}

func (s *Service) CreateGuestInvite(ctx context.Context, cmd CreateGuestInviteCmd) (*CreateGuestInviteResult, error) {
	access, channel, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	if !cmd.Actor.IsAccount() || !access.CanManage {
		return nil, fault.Forbidden("Only channel managers can create guest invites")
	}
	email, err := accdomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, fault.Validation("Invalid guest email")
	}
	rawToken, err := s.tokens.Generate()
	if err != nil {
		return nil, fault.Internal("Generate guest invite token failed", fault.WithCause(err))
	}
	canPost := true
	if cmd.CanPost != nil {
		canPost = *cmd.CanPost
	}
	now := s.now()
	invite, err := s.repo.CreateGuestInvite(ctx, collabdomain.GuestInvite{
		ID:             uuid.New(),
		ChannelID:      channel.ID,
		Email:          email.String(),
		CanPost:        canPost,
		VisibleFromSeq: channel.LastMessageSeq,
		ExpiresAt:      now.Add(s.guestInviteTTL),
		CreatedAt:      now,
		InvitedBy:      cmd.Actor.AccountID,
	}, s.tokens.Hash(rawToken))
	if err != nil {
		return nil, fault.Internal("Create guest invite failed", fault.WithCause(err))
	}
	return &CreateGuestInviteResult{Invite: invite, Token: rawToken, ExchangeURL: s.exchangeURL(rawToken)}, nil
}

func (s *Service) ExchangeGuestInvite(ctx context.Context, cmd ExchangeGuestInviteCmd) (*ExchangeGuestInviteResult, error) {
	displayName := strings.TrimSpace(cmd.DisplayName)
	if displayName == "" {
		return nil, fault.Validation("displayName is required")
	}
	token := strings.TrimSpace(cmd.Token)
	if token == "" {
		return nil, fault.Validation("invite token is required")
	}
	invite, err := s.repo.GetGuestInviteByTokenHash(ctx, s.tokens.Hash(token))
	if err != nil {
		return nil, fault.Internal("Load guest invite failed", fault.WithCause(err))
	}
	if invite == nil || invite.RevokedAt != nil || !invite.ExpiresAt.After(s.now()) {
		return nil, fault.Unauthorized("Guest invite is invalid or expired")
	}
	sessionRaw, err := s.tokens.Generate()
	if err != nil {
		return nil, fault.Internal("Generate guest session failed", fault.WithCause(err))
	}
	now := s.now()
	invite, guest, sessionID, err := s.repo.ExchangeGuestInvite(ctx, s.tokens.Hash(token), displayName, s.tokens.Hash(sessionRaw), cmd.UserAgent, cmd.IP, now.Add(s.guestAccessTTL), now.Add(s.guestAccessTTL), now)
	if err != nil {
		return nil, fault.Internal("Exchange guest invite failed", fault.WithCause(err))
	}
	accessToken, err := s.jwt.GenerateAccessToken(ctx, authdomain.NewGuestPrincipal(guest.ID, sessionID, guest.ChannelID), now.Add(s.jwt.AccessTTL()))
	if err != nil {
		return nil, fault.Internal("Generate guest access token failed", fault.WithCause(err))
	}
	return &ExchangeGuestInviteResult{Invite: invite, Guest: guest, AccessToken: accessToken, TokenType: "Bearer", ExpiresIn: int64(s.jwt.AccessTTL().Seconds())}, nil
}

func (s *Service) requireGroupAccountAccess(ctx context.Context, groupID uuid.UUID, principal authdomain.Principal) (collabdomain.Access, error) {
	if !principal.IsAccount() {
		return collabdomain.Access{}, fault.Unauthorized("Account authentication required")
	}
	access, err := s.repo.ResolveGroupAccessForAccount(ctx, groupID, principal.AccountID)
	if err != nil {
		return collabdomain.Access{}, fault.Internal("Resolve group access failed", fault.WithCause(err))
	}
	if !access.Allowed {
		return collabdomain.Access{}, fault.Forbidden("Group access denied")
	}
	return access, nil
}

func (s *Service) requireChannelAccess(ctx context.Context, channelID uuid.UUID, principal authdomain.Principal) (collabdomain.Access, *collabdomain.Channel, error) {
	channel, err := s.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		return collabdomain.Access{}, nil, fault.Internal("Load channel failed", fault.WithCause(err))
	}
	if channel == nil {
		return collabdomain.Access{}, nil, fault.NotFound("Channel not found")
	}
	var access collabdomain.Access
	switch {
	case principal.IsAccount():
		access, err = s.repo.ResolveChannelAccessForAccount(ctx, channelID, principal.AccountID)
	case principal.IsGuest():
		if principal.ChannelID != channelID {
			return collabdomain.Access{}, nil, fault.Forbidden("Guest access is scoped to a single channel")
		}
		access, err = s.repo.ResolveChannelAccessForGuest(ctx, channelID, principal.GuestID)
	default:
		return collabdomain.Access{}, nil, fault.Unauthorized("Authentication required")
	}
	if err != nil {
		return collabdomain.Access{}, nil, fault.Internal("Resolve channel access failed", fault.WithCause(err))
	}
	if !access.Allowed {
		return collabdomain.Access{}, nil, fault.Forbidden("Channel access denied")
	}
	return access, channel, nil
}

func (s *Service) canMutateMessage(access collabdomain.Access, principal authdomain.Principal, message *collabdomain.Message) bool {
	if access.CanModerate || access.CanManage {
		return true
	}
	if principal.IsAccount() && message.AuthorAccountID != nil {
		return *message.AuthorAccountID == principal.AccountID
	}
	if principal.IsGuest() && message.AuthorGuestID != nil {
		return *message.AuthorGuestID == principal.GuestID
	}
	return false
}

func buildAttachmentObjectKey(channelID, objectID uuid.UUID, fileName string, now time.Time) string {
	return strings.Join([]string{"collab", "channels", channelID.String(), now.UTC().Format("2006"), now.UTC().Format("01"), now.UTC().Format("02"), objectID.String(), sanitizeFileName(fileName)}, "/")
}

func sanitizeFileName(fileName string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "file.bin"
	}
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(strings.TrimSpace(b.String()), "-")
	if out == "" {
		return "file.bin"
	}
	return out
}

func (s *Service) exchangeURL(token string) string {
	path := "/api/v1/guest-invites/" + token + "/exchange"
	base := strings.TrimRight(strings.TrimSpace(s.publicBaseURL), "/")
	if base == "" {
		return path
	}
	return base + path
}

func (s *Service) AuthorizeChannel(ctx context.Context, channelID uuid.UUID, actor authdomain.Principal) error {
	_, _, err := s.requireChannelAccess(ctx, channelID, actor)
	return err
}
