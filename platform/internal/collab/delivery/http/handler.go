package http

import (
	"context"
	"strings"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	authapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	"github.com/NikolayNam/collabsphere/internal/collab/delivery/http/dto"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	svc *authapp.Service
}

func NewHandler(svc *authapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateChannel(ctx context.Context, input *dto.CreateChannelInput) (*dto.ChannelResponse, error) {
	groupID, err := parseUUID(input.GroupID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	adminIDs, err := parseUUIDSlice(input.Body.AdminAccountIDs)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	channel, err := h.svc.CreateChannel(ctx, authapp.CreateChannelCmd{GroupID: groupID, Actor: principal(ctx), Slug: input.Body.Slug, Name: input.Body.Name, Description: input.Body.Description, AdminAccountIDs: adminIDs})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.ChannelResponse{Status: 201, Body: toChannelPayload(*channel)}, nil
}

func (h *Handler) ListChannels(ctx context.Context, input *dto.ListChannelsInput) (*dto.ChannelsResponse, error) {
	groupID, err := parseUUID(input.GroupID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	items, err := h.svc.ListChannels(ctx, authapp.ListChannelsQuery{GroupID: groupID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.ChannelsResponse{Status: 200}
	resp.Body.Channels = make([]dto.ChannelPayload, 0, len(items))
	for _, item := range items {
		resp.Body.Channels = append(resp.Body.Channels, toChannelPayload(item))
	}
	return resp, nil
}

func (h *Handler) CreateMessage(ctx context.Context, input *dto.CreateMessageInput) (*dto.MessageResponse, error) {
	channelID, mentionIDs, attachmentIDs, replyToID, err := parseMessageInput(input.ChannelID, input.Body.MentionAccountIDs, input.Body.AttachmentObjectIDs, input.Body.ReplyToMessageID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	message, err := h.svc.CreateMessage(ctx, authapp.CreateMessageCmd{ChannelID: channelID, Actor: principal(ctx), Body: input.Body.Body, ReplyToMessageID: replyToID, MentionAccountIDs: mentionIDs, AttachmentObjectIDs: attachmentIDs})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.MessageResponse{Status: 201, Body: toMessagePayload(*message)}, nil
}

func (h *Handler) ListMessages(ctx context.Context, input *dto.ListMessagesInput) (*dto.MessagesResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	messages, err := h.svc.ListMessages(ctx, authapp.ListMessagesQuery{ChannelID: channelID, Actor: principal(ctx), Limit: input.Limit})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.MessagesResponse{Status: 200}
	resp.Body.Messages = make([]dto.MessagePayload, 0, len(messages))
	for _, message := range messages {
		resp.Body.Messages = append(resp.Body.Messages, toMessagePayload(message))
	}
	return resp, nil
}

func (h *Handler) UpdateMessage(ctx context.Context, input *dto.UpdateMessageInput) (*dto.MessageResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	messageID, err := parseUUID(input.MessageID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	mentionIDs, err := parseUUIDSlice(input.Body.MentionAccountIDs)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	message, err := h.svc.UpdateMessage(ctx, authapp.UpdateMessageCmd{ChannelID: channelID, MessageID: messageID, Actor: principal(ctx), Body: input.Body.Body, MentionAccountIDs: mentionIDs})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.MessageResponse{Status: 200, Body: toMessagePayload(*message)}, nil
}

func (h *Handler) DeleteMessage(ctx context.Context, input *dto.DeleteMessageInput) (*dto.EmptyResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	messageID, err := parseUUID(input.MessageID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	if err := h.svc.DeleteMessage(ctx, authapp.DeleteMessageCmd{ChannelID: channelID, MessageID: messageID, Actor: principal(ctx)}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
}

func (h *Handler) CreateAttachmentUpload(ctx context.Context, input *dto.CreateAttachmentUploadInput) (*dto.AttachmentUploadResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	orgID, err := parseOptionalUUID(input.Body.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	res, err := h.svc.CreateAttachmentUpload(ctx, authapp.CreateAttachmentUploadCmd{ChannelID: channelID, Actor: principal(ctx), OrganizationID: orgID, FileName: input.Body.FileName, ContentType: input.Body.ContentType, SizeBytes: input.Body.SizeBytes, ChecksumSHA256: input.Body.ChecksumSHA256})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AttachmentUploadResponse{Status: 201}
	out.Body.ObjectID = res.ObjectID
	out.Body.UploadURL = res.UploadURL
	out.Body.ExpiresAt = res.ExpiresAt
	out.Body.Bucket = res.Bucket
	out.Body.ObjectKey = res.ObjectKey
	out.Body.FileName = res.FileName
	out.Body.SizeBytes = res.SizeBytes
	return out, nil
}

func (h *Handler) UpdateReadCursor(ctx context.Context, input *dto.UpdateReadCursorInput) (*dto.EmptyResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	if err := h.svc.UpdateReadCursor(ctx, authapp.UpdateReadCursorCmd{ChannelID: channelID, Actor: principal(ctx), LastReadSeq: input.Body.LastReadSeq}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
}

func (h *Handler) ToggleReaction(ctx context.Context, input *dto.ToggleReactionInput) (*dto.EmptyResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	messageID, err := parseUUID(input.Body.MessageID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	if err := h.svc.ToggleReaction(ctx, authapp.ToggleReactionCmd{ChannelID: channelID, MessageID: messageID, Actor: principal(ctx), Emoji: input.Body.Emoji}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
}

func (h *Handler) CreateGuestInvite(ctx context.Context, input *dto.CreateGuestInviteInput) (*dto.GuestInviteResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	res, err := h.svc.CreateGuestInvite(ctx, authapp.CreateGuestInviteCmd{ChannelID: channelID, Actor: principal(ctx), Email: input.Body.Email, CanPost: input.Body.CanPost})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.GuestInviteResponse{Status: 201}
	out.Body.Invite = toGuestInvitePayload(*res.Invite)
	out.Body.Token = res.Token
	out.Body.ExchangeURL = res.ExchangeURL
	return out, nil
}

func (h *Handler) ExchangeGuestInvite(ctx context.Context, input *dto.ExchangeGuestInviteInput) (*dto.ExchangeGuestInviteResponse, error) {
	res, err := h.svc.ExchangeGuestInvite(ctx, authapp.ExchangeGuestInviteCmd{Token: input.Token, DisplayName: input.Body.DisplayName, UserAgent: optionalString(input.UserAgent), IP: extractClientIP(input.XForwardedFor, input.XRealIP)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.ExchangeGuestInviteResponse{Status: 200}
	out.Body.Invite = toGuestInvitePayload(*res.Invite)
	out.Body.Guest = toGuestIdentityPayload(*res.Guest)
	out.Body.AccessToken = res.AccessToken
	out.Body.TokenType = res.TokenType
	out.Body.ExpiresIn = res.ExpiresIn
	return out, nil
}

func (h *Handler) CreateConference(ctx context.Context, input *dto.CreateConferenceInput) (*dto.ConferenceResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	conference, err := h.svc.CreateConference(ctx, authapp.CreateConferenceCmd{ChannelID: channelID, Actor: principal(ctx), Kind: input.Body.Kind, Title: input.Body.Title, ScheduledStartAt: input.Body.ScheduledStartAt, RecordingEnabled: input.Body.RecordingEnabled})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.ConferenceResponse{Status: 201, Body: toConferencePayload(*conference)}, nil
}

func (h *Handler) ListConferences(ctx context.Context, input *dto.ListConferencesInput) (*dto.ConferencesResponse, error) {
	channelID, err := parseUUID(input.ChannelID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	items, err := h.svc.ListConferences(ctx, authapp.ListConferencesQuery{ChannelID: channelID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.ConferencesResponse{Status: 200}
	out.Body.Conferences = make([]dto.ConferencePayload, 0, len(items))
	for _, item := range items {
		out.Body.Conferences = append(out.Body.Conferences, toConferencePayload(item))
	}
	return out, nil
}

func (h *Handler) CreateConferenceJoinToken(ctx context.Context, input *dto.CreateConferenceJoinTokenInput) (*dto.ConferenceJoinTokenResponse, error) {
	conferenceID, err := parseUUID(input.ConferenceID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	res, err := h.svc.CreateConferenceJoinToken(ctx, authapp.CreateConferenceJoinTokenCmd{ConferenceID: conferenceID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.ConferenceJoinTokenResponse{Status: 200}
	out.Body.Token = res.Token
	out.Body.RoomName = res.RoomName
	out.Body.JoinURL = res.JoinURL
	out.Body.ExpiresAt = res.ExpiresAt
	return out, nil
}

func (h *Handler) StartConferenceRecording(ctx context.Context, input *dto.UpdateConferenceRecordingInput) (*dto.ConferenceResponse, error) {
	conferenceID, err := parseUUID(input.ConferenceID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	conference, err := h.svc.StartConferenceRecording(ctx, authapp.UpdateConferenceRecordingCmd{ConferenceID: conferenceID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.ConferenceResponse{Status: 200, Body: toConferencePayload(*conference)}, nil
}

func (h *Handler) StopConferenceRecording(ctx context.Context, input *dto.UpdateConferenceRecordingInput) (*dto.ConferenceResponse, error) {
	conferenceID, err := parseUUID(input.ConferenceID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	conference, err := h.svc.StopConferenceRecording(ctx, authapp.UpdateConferenceRecordingCmd{ConferenceID: conferenceID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.ConferenceResponse{Status: 200, Body: toConferencePayload(*conference)}, nil
}

func (h *Handler) GetConferenceTranscript(ctx context.Context, input *dto.GetConferenceTranscriptInput) (*dto.ConferenceTranscriptResponse, error) {
	conferenceID, err := parseUUID(input.ConferenceID)
	if err != nil {
		return nil, humaerr.From(ctx, authappValidation())
	}
	transcript, err := h.svc.GetConferenceTranscript(ctx, authapp.GetConferenceTranscriptQuery{ConferenceID: conferenceID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.ConferenceTranscriptResponse{Status: 200}
	out.Body.ConferenceID = transcript.ConferenceID
	out.Body.TranscriptText = transcript.TranscriptText
	out.Body.SegmentsJSON = transcript.SegmentsJSON
	out.Body.LanguageCode = transcript.LanguageCode
	out.Body.SourceRecordingID = transcript.SourceRecordingID
	out.Body.CreatedAt = transcript.CreatedAt
	out.Body.UpdatedAt = transcript.UpdatedAt
	return out, nil
}

func (h *Handler) JitsiWebhook(ctx context.Context, input *dto.JitsiWebhookInput) (*dto.EmptyResponse, error) {
	if err := h.svc.HandleJitsiWebhook(ctx, authapp.JitsiWebhookCmd{ProviderEventID: input.Body.ProviderEventID, EventType: input.Body.EventType, Payload: input.Body.Payload}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: 204}, nil
}

func principal(ctx context.Context) authdomain.Principal {
	return authmw.PrincipalFromContext(ctx)
}

func authappValidation() error {
	return fault.Validation("Invalid identifier")
}

func parseMessageInput(channelID string, mentions []string, attachments []string, replyTo *string) (uuid.UUID, []uuid.UUID, []uuid.UUID, *uuid.UUID, error) {
	parsedChannelID, err := parseUUID(channelID)
	if err != nil {
		return uuid.Nil, nil, nil, nil, err
	}
	parsedMentions, err := parseUUIDSlice(mentions)
	if err != nil {
		return uuid.Nil, nil, nil, nil, err
	}
	parsedAttachments, err := parseUUIDSlice(attachments)
	if err != nil {
		return uuid.Nil, nil, nil, nil, err
	}
	parsedReplyTo, err := parseOptionalUUID(replyTo)
	if err != nil {
		return uuid.Nil, nil, nil, nil, err
	}
	return parsedChannelID, parsedMentions, parsedAttachments, parsedReplyTo, nil
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errInvalidID
	}
	return id, nil
}

func parseUUIDSlice(values []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := parseUUID(value)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := parseUUID(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

var errInvalidID = fault.Validation("Invalid identifier")

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func extractClientIP(forwardedFor, realIP string) *string {
	if value := firstHeaderValue(forwardedFor); value != "" {
		return &value
	}
	if value := strings.TrimSpace(realIP); value != "" {
		return &value
	}
	return nil
}

func firstHeaderValue(value string) string {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}

func toChannelPayload(channel collabdomain.Channel) dto.ChannelPayload {
	return dto.ChannelPayload{ID: channel.ID, GroupID: channel.GroupID, Slug: channel.Slug, Name: channel.Name, Description: channel.Description, IsDefault: channel.IsDefault, LastMessageSeq: channel.LastMessageSeq, CreatedAt: channel.CreatedAt, UpdatedAt: channel.UpdatedAt}
}

func toMessagePayload(message collabdomain.Message) dto.MessagePayload {
	attachments := make([]dto.AttachmentPayload, 0, len(message.Attachments))
	for _, attachment := range message.Attachments {
		attachments = append(attachments, dto.AttachmentPayload{ObjectID: attachment.ObjectID, OrganizationID: attachment.OrganizationID, FileName: attachment.FileName, ContentType: attachment.ContentType, SizeBytes: attachment.SizeBytes, Bucket: attachment.Bucket, ObjectKey: attachment.ObjectKey, CreatedAt: attachment.CreatedAt})
	}
	reactions := make([]dto.ReactionPayload, 0, len(message.Reactions))
	for _, reaction := range message.Reactions {
		reactions = append(reactions, dto.ReactionPayload{Emoji: reaction.Emoji, Count: reaction.Count, Mine: reaction.Mine})
	}
	return dto.MessagePayload{ID: message.ID, ChannelID: message.ChannelID, ChannelSeq: message.ChannelSeq, Type: string(message.Type), AuthorType: string(message.AuthorType), AuthorAccountID: message.AuthorAccountID, AuthorGuestID: message.AuthorGuestID, AuthorName: message.AuthorName, Body: message.Body, ReplyToMessageID: message.ReplyToMessageID, CreatedAt: message.CreatedAt, EditedAt: message.EditedAt, DeletedAt: message.DeletedAt, Mentions: message.Mentions, Attachments: attachments, Reactions: reactions}
}

func toGuestInvitePayload(invite collabdomain.GuestInvite) dto.GuestInvitePayload {
	return dto.GuestInvitePayload{ID: invite.ID, ChannelID: invite.ChannelID, Email: invite.Email, CanPost: invite.CanPost, VisibleFromSeq: invite.VisibleFromSeq, ExpiresAt: invite.ExpiresAt, AcceptedAt: invite.AcceptedAt, RevokedAt: invite.RevokedAt, CreatedAt: invite.CreatedAt, InvitedBy: invite.InvitedBy}
}

func toGuestIdentityPayload(guest collabdomain.GuestIdentity) dto.GuestIdentityPayload {
	return dto.GuestIdentityPayload{ID: guest.ID, InviteID: guest.InviteID, ChannelID: guest.ChannelID, Email: guest.Email, DisplayName: guest.DisplayName, ExpiresAt: guest.ExpiresAt, CreatedAt: guest.CreatedAt}
}

func toConferencePayload(conference collabdomain.Conference) dto.ConferencePayload {
	return dto.ConferencePayload{ID: conference.ID, ChannelID: conference.ChannelID, Kind: string(conference.Kind), Status: string(conference.Status), Provider: conference.Provider, Title: conference.Title, JitsiRoomName: conference.JitsiRoomName, ScheduledStartAt: conference.ScheduledStartAt, StartedAt: conference.StartedAt, EndedAt: conference.EndedAt, RecordingEnabled: conference.RecordingEnabled, RecordingStartedAt: conference.RecordingStartedAt, RecordingStoppedAt: conference.RecordingStoppedAt, TranscriptionStatus: string(conference.TranscriptionStatus), CreatedAt: conference.CreatedAt, UpdatedAt: conference.UpdatedAt}
}
