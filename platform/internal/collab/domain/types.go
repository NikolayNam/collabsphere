package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ActorType string

type MessageType string

type ConferenceKind string

type ConferenceStatus string

type TranscriptionStatus string

const (
	ActorTypeAccount ActorType = "account"
	ActorTypeGuest   ActorType = "guest"
	ActorTypeSystem  ActorType = "system"

	MessageTypeUser   MessageType = "user"
	MessageTypeSystem MessageType = "system"

	ConferenceKindAudio ConferenceKind = "audio"
	ConferenceKindVideo ConferenceKind = "video"

	ConferenceStatusScheduled ConferenceStatus = "scheduled"
	ConferenceStatusLive      ConferenceStatus = "live"
	ConferenceStatusEnded     ConferenceStatus = "ended"
	ConferenceStatusCancelled ConferenceStatus = "cancelled"

	TranscriptionStatusPending    TranscriptionStatus = "pending"
	TranscriptionStatusProcessing TranscriptionStatus = "processing"
	TranscriptionStatusReady      TranscriptionStatus = "ready"
	TranscriptionStatusFailed     TranscriptionStatus = "failed"
	TranscriptionStatusDisabled   TranscriptionStatus = "disabled"
)

type Access struct {
	GroupID         uuid.UUID
	ChannelID       uuid.UUID
	Allowed         bool
	CanRead         bool
	CanPost         bool
	CanManage       bool
	CanModerate     bool
	IsGuest         bool
	VisibleFromSeq  int64
	GroupRole       string
	ChannelAdmin    bool
	InviteID        uuid.UUID
	OrganizationIDs []uuid.UUID
}

type Channel struct {
	ID             uuid.UUID
	GroupID        uuid.UUID
	Slug           string
	Name           string
	Description    *string
	IsDefault      bool
	LastMessageSeq int64
	CreatedBy      *uuid.UUID
	UpdatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

type Attachment struct {
	ObjectID       uuid.UUID
	OrganizationID *uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	Bucket         string
	ObjectKey      string
	CreatedAt      time.Time
}

type ReactionSummary struct {
	Emoji string
	Count int64
	Mine  bool
}

type Message struct {
	ID               uuid.UUID
	ChannelID        uuid.UUID
	ChannelSeq       int64
	Type             MessageType
	AuthorType       ActorType
	AuthorAccountID  *uuid.UUID
	AuthorGuestID    *uuid.UUID
	AuthorName       *string
	Body             string
	ReplyToMessageID *uuid.UUID
	CreatedAt        time.Time
	EditedAt         *time.Time
	DeletedAt        *time.Time
	Mentions         []uuid.UUID
	Attachments      []Attachment
	Reactions        []ReactionSummary
}

type GuestInvite struct {
	ID             uuid.UUID
	ChannelID      uuid.UUID
	Email          string
	CanPost        bool
	VisibleFromSeq int64
	ExpiresAt      time.Time
	AcceptedAt     *time.Time
	RevokedAt      *time.Time
	CreatedAt      time.Time
	InvitedBy      uuid.UUID
}

type GuestIdentity struct {
	ID          uuid.UUID
	InviteID    uuid.UUID
	ChannelID   uuid.UUID
	Email       string
	DisplayName string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type Conference struct {
	ID                  uuid.UUID
	ChannelID           uuid.UUID
	Kind                ConferenceKind
	Status              ConferenceStatus
	Provider            string
	Title               string
	RoomName            string
	ScheduledStartAt    *time.Time
	StartedAt           *time.Time
	EndedAt             *time.Time
	RecordingEnabled    bool
	RecordingStartedAt  *time.Time
	RecordingStoppedAt  *time.Time
	TranscriptionStatus TranscriptionStatus
	CreatedBy           *uuid.UUID
	UpdatedBy           *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}

type ConferenceRecording struct {
	ID           uuid.UUID
	ConferenceID uuid.UUID
	ObjectID     uuid.UUID
	DurationSec  *int32
	MimeType     *string
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	FileName     *string
	Bucket       *string
	ObjectKey    *string
	ContentType  *string
	SizeBytes    *int64
}

type ConferenceTranscript struct {
	ConferenceID      uuid.UUID
	TranscriptText    string
	SegmentsJSON      json.RawMessage
	LanguageCode      *string
	SourceRecordingID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

type StorageObject struct {
	ID             uuid.UUID
	OrganizationID *uuid.UUID
	Bucket         string
	ObjectKey      string
	FileName       string
	ContentType    *string
	SizeBytes      int64
	ChecksumSHA256 *string
	CreatedAt      time.Time
}

type TranscriptionJob struct {
	ID           uuid.UUID
	ConferenceID uuid.UUID
	RecordingID  uuid.UUID
	Attempts     int
}

type Event struct {
	Type      string      `json:"type"`
	ChannelID uuid.UUID   `json:"channelId"`
	Payload   interface{} `json:"payload,omitempty"`
}
