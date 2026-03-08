package application

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	"github.com/google/uuid"
)

type AccountReader interface {
	GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error)
}

type ObjectStorage interface {
	PresignPutObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error)
	ReadObject(ctx context.Context, bucket, objectKey string) (io.ReadCloser, error)
}

type TokenGenerator interface {
	Generate() (string, error)
	Hash(raw string) string
}

type AccessTokenManager interface {
	GenerateAccessToken(ctx context.Context, principal authdomain.Principal, expiresAt time.Time) (string, error)
	AccessTTL() time.Duration
}

type EventPublisher interface {
	Publish(ctx context.Context, event collabdomain.Event)
}

type Clock interface {
	Now() time.Time
}

type Transcriber interface {
	Transcribe(ctx context.Context, fileName string, mimeType *string, content io.Reader) (TranscriptionResult, error)
}

type TranscriptionResult struct {
	Text         string
	SegmentsJSON json.RawMessage
	LanguageCode *string
}

type Service struct {
	repo               *collabpg.Repo
	accounts           AccountReader
	storage            ObjectStorage
	tokens             TokenGenerator
	jwt                AccessTokenManager
	clock              Clock
	publisher          EventPublisher
	transcriber        Transcriber
	conferenceProvider string
	publicBaseURL      string
	storageBucket      string
	guestInviteTTL     time.Duration
	guestAccessTTL     time.Duration
}

func New(repo *collabpg.Repo, accounts AccountReader, storage ObjectStorage, tokens TokenGenerator, jwt AccessTokenManager, clock Clock, publisher EventPublisher, transcriber Transcriber, conferenceProvider, publicBaseURL, storageBucket string, guestInviteTTL, guestAccessTTL time.Duration) *Service {
	provider := strings.ToLower(strings.TrimSpace(conferenceProvider))
	if provider == "" {
		provider = "mediasoup"
	}
	return &Service{
		repo:               repo,
		accounts:           accounts,
		storage:            storage,
		tokens:             tokens,
		jwt:                jwt,
		clock:              clock,
		publisher:          publisher,
		transcriber:        transcriber,
		conferenceProvider: provider,
		publicBaseURL:      publicBaseURL,
		storageBucket:      storageBucket,
		guestInviteTTL:     guestInviteTTL,
		guestAccessTTL:     guestAccessTTL,
	}
}

func (s *Service) ConferenceProvider() string {
	if strings.TrimSpace(s.conferenceProvider) == "" {
		return "mediasoup"
	}
	return s.conferenceProvider
}

type CreateChannelCmd struct {
	GroupID         uuid.UUID
	Actor           authdomain.Principal
	Slug            string
	Name            string
	Description     *string
	AdminAccountIDs []uuid.UUID
}

type ListChannelsQuery struct {
	GroupID uuid.UUID
	Actor   authdomain.Principal
}

type CreateMessageCmd struct {
	ChannelID           uuid.UUID
	Actor               authdomain.Principal
	Body                string
	ReplyToMessageID    *uuid.UUID
	MentionAccountIDs   []uuid.UUID
	AttachmentObjectIDs []uuid.UUID
}

type UpdateMessageCmd struct {
	ChannelID         uuid.UUID
	MessageID         uuid.UUID
	Actor             authdomain.Principal
	Body              string
	MentionAccountIDs []uuid.UUID
}

type DeleteMessageCmd struct {
	ChannelID uuid.UUID
	MessageID uuid.UUID
	Actor     authdomain.Principal
}

type ListMessagesQuery struct {
	ChannelID uuid.UUID
	Actor     authdomain.Principal
	Limit     int
}

type CreateAttachmentUploadCmd struct {
	ChannelID      uuid.UUID
	Actor          authdomain.Principal
	OrganizationID *uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	ChecksumSHA256 *string
}

type CreateAttachmentUploadResult struct {
	ObjectID  uuid.UUID
	UploadURL string
	ExpiresAt time.Time
	Bucket    string
	ObjectKey string
	FileName  string
	SizeBytes int64
}

type UpdateReadCursorCmd struct {
	ChannelID   uuid.UUID
	Actor       authdomain.Principal
	LastReadSeq int64
}

type ToggleReactionCmd struct {
	ChannelID uuid.UUID
	MessageID uuid.UUID
	Actor     authdomain.Principal
	Emoji     string
}

type CreateGuestInviteCmd struct {
	ChannelID uuid.UUID
	Actor     authdomain.Principal
	Email     string
	CanPost   *bool
}

type CreateGuestInviteResult struct {
	Invite      *collabdomain.GuestInvite
	Token       string
	ExchangeURL string
}

type ExchangeGuestInviteCmd struct {
	Token       string
	DisplayName string
	UserAgent   *string
	IP          *string
}

type ExchangeGuestInviteResult struct {
	Invite      *collabdomain.GuestInvite
	Guest       *collabdomain.GuestIdentity
	AccessToken string
	TokenType   string
	ExpiresIn   int64
}

type CreateConferenceCmd struct {
	ChannelID        uuid.UUID
	Actor            authdomain.Principal
	Kind             string
	Title            string
	ScheduledStartAt *time.Time
	RecordingEnabled bool
}

type ListConferencesQuery struct {
	ChannelID uuid.UUID
	Actor     authdomain.Principal
}

type CreateConferenceJoinTokenCmd struct {
	ConferenceID uuid.UUID
	Actor        authdomain.Principal
}

type CreateConferenceJoinTokenResult struct {
	Token     string
	RoomName  string
	JoinURL   string
	ExpiresAt time.Time
}

type UpdateConferenceRecordingCmd struct {
	ConferenceID uuid.UUID
	Actor        authdomain.Principal
}

type GetConferenceTranscriptQuery struct {
	ConferenceID uuid.UUID
	Actor        authdomain.Principal
}
