package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EmptyResponse struct {
	Status int
}

type ChannelPayload struct {
	ID             uuid.UUID  `json:"id"`
	GroupID        uuid.UUID  `json:"groupId"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	IsDefault      bool       `json:"isDefault"`
	LastMessageSeq int64      `json:"lastMessageSeq"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

type ChannelsResponse struct {
	Status int
	Body   struct {
		Channels []ChannelPayload `json:"channels"`
	}
}

type ChannelResponse struct {
	Status int
	Body   ChannelPayload
}

type CreateChannelInput struct {
	GroupID string `path:"group_id"`
	Body    struct {
		Slug            string   `json:"slug"`
		Name            string   `json:"name"`
		Description     *string  `json:"description,omitempty"`
		AdminAccountIDs []string `json:"adminAccountIds,omitempty"`
	}
}

type ListChannelsInput struct {
	GroupID string `path:"group_id"`
}

type AttachmentPayload struct {
	ObjectID       uuid.UUID  `json:"objectId"`
	OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
	FileName       string     `json:"fileName"`
	ContentType    *string    `json:"contentType,omitempty"`
	SizeBytes      int64      `json:"sizeBytes"`
	Bucket         string     `json:"bucket"`
	ObjectKey      string     `json:"objectKey"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type ReactionPayload struct {
	Emoji string `json:"emoji"`
	Count int64  `json:"count"`
	Mine  bool   `json:"mine"`
}

type MessagePayload struct {
	ID               uuid.UUID           `json:"id"`
	ChannelID        uuid.UUID           `json:"channelId"`
	ChannelSeq       int64               `json:"channelSeq"`
	Type             string              `json:"type"`
	AuthorType       string              `json:"authorType"`
	AuthorAccountID  *uuid.UUID          `json:"authorAccountId,omitempty"`
	AuthorGuestID    *uuid.UUID          `json:"authorGuestId,omitempty"`
	AuthorName       *string             `json:"authorName,omitempty"`
	Body             string              `json:"body"`
	ReplyToMessageID *uuid.UUID          `json:"replyToMessageId,omitempty"`
	CreatedAt        time.Time           `json:"createdAt"`
	EditedAt         *time.Time          `json:"editedAt,omitempty"`
	DeletedAt        *time.Time          `json:"deletedAt,omitempty"`
	Mentions         []uuid.UUID         `json:"mentions,omitempty"`
	Attachments      []AttachmentPayload `json:"attachments,omitempty"`
	Reactions        []ReactionPayload   `json:"reactions,omitempty"`
}

type MessagesResponse struct {
	Status int
	Body   struct {
		Messages []MessagePayload `json:"messages"`
	}
}

type MessageResponse struct {
	Status int
	Body   MessagePayload
}

type ListMessagesInput struct {
	ChannelID string `path:"channel_id"`
	Limit     int    `query:"limit"`
}

type CreateMessageInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		Body                string   `json:"body"`
		ReplyToMessageID    *string  `json:"replyToMessageId,omitempty"`
		MentionAccountIDs   []string `json:"mentionAccountIds,omitempty"`
		AttachmentObjectIDs []string `json:"attachmentObjectIds,omitempty"`
	}
}

type UpdateMessageInput struct {
	ChannelID string `path:"channel_id"`
	MessageID string `path:"message_id"`
	Body      struct {
		Body              string   `json:"body"`
		MentionAccountIDs []string `json:"mentionAccountIds,omitempty"`
	}
}

type DeleteMessageInput struct {
	ChannelID string `path:"channel_id"`
	MessageID string `path:"message_id"`
}

type CreateAttachmentUploadInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		OrganizationID *string `json:"organizationId,omitempty"`
		FileName       string  `json:"fileName"`
		ContentType    *string `json:"contentType,omitempty"`
		SizeBytes      int64   `json:"sizeBytes"`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty"`
	}
}

type AttachmentUploadResponse struct {
	Status int
	Body   struct {
		ObjectID  uuid.UUID `json:"objectId"`
		UploadURL string    `json:"uploadUrl"`
		ExpiresAt time.Time `json:"expiresAt"`
		Bucket    string    `json:"bucket"`
		ObjectKey string    `json:"objectKey"`
		FileName  string    `json:"fileName"`
		SizeBytes int64     `json:"sizeBytes"`
	}
}

type UpdateReadCursorInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		LastReadSeq int64 `json:"lastReadSeq"`
	}
}

type ToggleReactionInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		MessageID string `json:"messageId"`
		Emoji     string `json:"emoji"`
	}
}

type CreateGuestInviteInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		Email   string `json:"email"`
		CanPost *bool  `json:"canPost,omitempty"`
	}
}

type GuestInvitePayload struct {
	ID             uuid.UUID  `json:"id"`
	ChannelID      uuid.UUID  `json:"channelId"`
	Email          string     `json:"email"`
	CanPost        bool       `json:"canPost"`
	VisibleFromSeq int64      `json:"visibleFromSeq"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	InvitedBy      uuid.UUID  `json:"invitedBy"`
}

type GuestInviteResponse struct {
	Status int
	Body   struct {
		Invite      GuestInvitePayload `json:"invite"`
		Token       string             `json:"token"`
		ExchangeURL string             `json:"exchangeUrl"`
	}
}

type ExchangeGuestInviteInput struct {
	Token string `path:"token"`
	Body  struct {
		DisplayName string `json:"displayName"`
	}
	UserAgent     string `header:"User-Agent"`
	XForwardedFor string `header:"X-Forwarded-For"`
	XRealIP       string `header:"X-Real-IP"`
}

type GuestIdentityPayload struct {
	ID          uuid.UUID `json:"id"`
	InviteID    uuid.UUID `json:"inviteId"`
	ChannelID   uuid.UUID `json:"channelId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	ExpiresAt   time.Time `json:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ExchangeGuestInviteResponse struct {
	Status int
	Body   struct {
		Invite      GuestInvitePayload   `json:"invite"`
		Guest       GuestIdentityPayload `json:"guest"`
		AccessToken string               `json:"accessToken"`
		TokenType   string               `json:"tokenType"`
		ExpiresIn   int64                `json:"expiresIn"`
	}
}

type ConferencePayload struct {
	ID                  uuid.UUID  `json:"id"`
	ChannelID           uuid.UUID  `json:"channelId"`
	Kind                string     `json:"kind"`
	Status              string     `json:"status"`
	Provider            string     `json:"provider"`
	Title               string     `json:"title"`
	JitsiRoomName       string     `json:"jitsiRoomName"`
	ScheduledStartAt    *time.Time `json:"scheduledStartAt,omitempty"`
	StartedAt           *time.Time `json:"startedAt,omitempty"`
	EndedAt             *time.Time `json:"endedAt,omitempty"`
	RecordingEnabled    bool       `json:"recordingEnabled"`
	RecordingStartedAt  *time.Time `json:"recordingStartedAt,omitempty"`
	RecordingStoppedAt  *time.Time `json:"recordingStoppedAt,omitempty"`
	TranscriptionStatus string     `json:"transcriptionStatus"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
}

type ConferenceResponse struct {
	Status int
	Body   ConferencePayload
}

type ConferencesResponse struct {
	Status int
	Body   struct {
		Conferences []ConferencePayload `json:"conferences"`
	}
}

type CreateConferenceInput struct {
	ChannelID string `path:"channel_id"`
	Body      struct {
		Kind             string     `json:"kind"`
		Title            string     `json:"title"`
		ScheduledStartAt *time.Time `json:"scheduledStartAt,omitempty"`
		RecordingEnabled bool       `json:"recordingEnabled"`
	}
}

type ListConferencesInput struct {
	ChannelID string `path:"channel_id"`
}

type CreateConferenceJoinTokenInput struct {
	ConferenceID string `path:"conference_id"`
}

type ConferenceJoinTokenResponse struct {
	Status int
	Body   struct {
		Token     string    `json:"token"`
		RoomName  string    `json:"roomName"`
		JoinURL   string    `json:"joinUrl"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
}

type UpdateConferenceRecordingInput struct {
	ConferenceID string `path:"conference_id"`
}

type GetConferenceTranscriptInput struct {
	ConferenceID string `path:"conference_id"`
}

type ConferenceTranscriptResponse struct {
	Status int
	Body   struct {
		ConferenceID      uuid.UUID       `json:"conferenceId"`
		TranscriptText    string          `json:"transcriptText"`
		SegmentsJSON      json.RawMessage `json:"segmentsJson"`
		LanguageCode      *string         `json:"languageCode,omitempty"`
		SourceRecordingID *uuid.UUID      `json:"sourceRecordingId,omitempty"`
		CreatedAt         time.Time       `json:"createdAt"`
		UpdatedAt         *time.Time      `json:"updatedAt,omitempty"`
	}
}

type JitsiWebhookInput struct {
	Body struct {
		ProviderEventID string          `json:"providerEventId"`
		EventType       string          `json:"eventType"`
		Payload         json.RawMessage `json:"payload"`
	}
}
