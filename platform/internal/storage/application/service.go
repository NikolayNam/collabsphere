package application

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

type ObjectStorage interface {
	PresignGetObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error)
}

type Repository interface {
	GetObjectByID(ctx context.Context, objectID uuid.UUID) (*StoredObject, error)
	GetAccountAvatarObjectID(ctx context.Context, accountID uuid.UUID) (*uuid.UUID, error)
	GetAccountVideoObjectID(ctx context.Context, accountID, videoID uuid.UUID) (*uuid.UUID, error)
	GetOrganizationLogoObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error)
	GetOrganizationVideoObjectID(ctx context.Context, organizationID, videoID uuid.UUID) (*uuid.UUID, error)
	GetCooperationPriceListObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error)
	GetOrganizationLegalDocumentObjectID(ctx context.Context, organizationID, documentID uuid.UUID) (*uuid.UUID, error)
	GetProductImportSourceObjectID(ctx context.Context, organizationID, batchID uuid.UUID) (*uuid.UUID, error)
	GetProductVideoObjectID(ctx context.Context, organizationID, productID, videoID uuid.UUID) (*uuid.UUID, error)
	GetConferenceChannelID(ctx context.Context, conferenceID uuid.UUID) (*uuid.UUID, error)
	GetConferenceRecordingObjectID(ctx context.Context, conferenceID, recordingID uuid.UUID) (*uuid.UUID, error)
	ListConferenceRecordings(ctx context.Context, conferenceID uuid.UUID) ([]ConferenceRecordingFile, error)
	ChannelHasAttachmentObject(ctx context.Context, channelID, objectID uuid.UUID) (bool, error)
	AccountOwnsAvatar(ctx context.Context, accountID, objectID uuid.UUID) (bool, error)
	AccountOwnsVideo(ctx context.Context, accountID, objectID uuid.UUID) (bool, error)
	ListRelatedOrganizationIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error)
	AccountHasAnyOrganizationAccess(ctx context.Context, accountID uuid.UUID, organizationIDs []uuid.UUID) (bool, error)
	ListRelatedChannelIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error)
	ListAccountFiles(ctx context.Context, accountID uuid.UUID) ([]ListedFile, error)
	ListOrganizationFiles(ctx context.Context, organizationID uuid.UUID) ([]ListedFile, error)
}

type ChannelAccessResolver interface {
	ResolveChannelAccessForAccount(ctx context.Context, channelID, accountID uuid.UUID) (collabdomain.Access, error)
	ResolveChannelAccessForGuest(ctx context.Context, channelID, guestID uuid.UUID) (collabdomain.Access, error)
}

type StoredObject struct {
	ID             uuid.UUID
	Bucket         string
	ObjectKey      string
	FileName       string
	ContentType    *string
	SizeBytes      int64
	OrganizationID *uuid.UUID
	CreatedAt      time.Time
}

type ListedFile struct {
	ObjectID       uuid.UUID
	OrganizationID *uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	CreatedAt      time.Time
	SourceType     string
	SourceID       *uuid.UUID
}

type ConferenceRecordingFile struct {
	RecordingID  uuid.UUID
	ConferenceID uuid.UUID
	ObjectID     uuid.UUID
	FileName     string
	ContentType  *string
	SizeBytes    int64
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	DurationSec  *int32
	MimeType     *string
}

type DownloadObjectQuery struct {
	ObjectID uuid.UUID
	Actor    authdomain.Principal
}

type DownloadObjectResult struct {
	ObjectID       uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	DownloadURL    string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	OrganizationID *uuid.UUID
}

type ListMyFilesQuery struct {
	Actor authdomain.Principal
}

type ListOrganizationFilesQuery struct {
	OrganizationID uuid.UUID
	Actor          authdomain.Principal
}

type DownloadMyAvatarQuery struct {
	Actor authdomain.Principal
}

type DownloadMyAccountVideoQuery struct {
	VideoID uuid.UUID
	Actor   authdomain.Principal
}

type DownloadOrganizationLogoQuery struct {
	OrganizationID uuid.UUID
	Actor          authdomain.Principal
}

type DownloadOrganizationVideoQuery struct {
	OrganizationID uuid.UUID
	VideoID        uuid.UUID
	Actor          authdomain.Principal
}

type DownloadCooperationPriceListQuery struct {
	OrganizationID uuid.UUID
	Actor          authdomain.Principal
}

type DownloadOrganizationLegalDocumentQuery struct {
	OrganizationID uuid.UUID
	DocumentID     uuid.UUID
	Actor          authdomain.Principal
}

type DownloadProductImportSourceQuery struct {
	OrganizationID uuid.UUID
	BatchID        uuid.UUID
	Actor          authdomain.Principal
}

type DownloadProductVideoQuery struct {
	OrganizationID uuid.UUID
	ProductID      uuid.UUID
	VideoID        uuid.UUID
	Actor          authdomain.Principal
}

type DownloadChannelAttachmentQuery struct {
	ChannelID uuid.UUID
	ObjectID  uuid.UUID
	Actor     authdomain.Principal
}

type DownloadConferenceRecordingQuery struct {
	ConferenceID uuid.UUID
	RecordingID  uuid.UUID
	Actor        authdomain.Principal
}

type ListConferenceRecordingsQuery struct {
	ConferenceID uuid.UUID
	Actor        authdomain.Principal
}

type Service struct {
	repo        Repository
	memberships memberports.MembershipRepository
	channels    ChannelAccessResolver
	storage     ObjectStorage
}

func New(repo Repository, memberships memberports.MembershipRepository, channels ChannelAccessResolver, storage ObjectStorage) *Service {
	return &Service{repo: repo, memberships: memberships, channels: channels, storage: storage}
}

func (s *Service) CreateDownload(ctx context.Context, q DownloadObjectQuery) (*DownloadObjectResult, error) {
	if q.ObjectID == uuid.Nil {
		return nil, fault.Validation("Object ID is required")
	}
	if !q.Actor.Authenticated {
		return nil, fault.Unauthorized("Authentication required")
	}
	if s.storage == nil {
		return nil, fault.Unavailable("File download is unavailable")
	}

	obj, err := s.repo.GetObjectByID(ctx, q.ObjectID)
	if err != nil {
		return nil, fault.Internal("Load storage object failed", fault.WithCause(err))
	}
	if obj == nil {
		return nil, fault.NotFound("Storage object not found")
	}

	allowed, err := s.canDownload(ctx, obj, q.Actor)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fault.Forbidden("Storage object access denied")
	}

	url, expiresAt, err := s.storage.PresignGetObject(ctx, obj.Bucket, obj.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign file download failed", fault.WithCause(err))
	}

	return &DownloadObjectResult{
		ObjectID:       obj.ID,
		FileName:       obj.FileName,
		ContentType:    obj.ContentType,
		SizeBytes:      obj.SizeBytes,
		DownloadURL:    url,
		ExpiresAt:      expiresAt,
		CreatedAt:      obj.CreatedAt,
		OrganizationID: obj.OrganizationID,
	}, nil
}

func (s *Service) ListMyFiles(ctx context.Context, q ListMyFilesQuery) ([]ListedFile, error) {
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	files, err := s.repo.ListAccountFiles(ctx, q.Actor.AccountID)
	if err != nil {
		return nil, fault.Internal("List account files failed", fault.WithCause(err))
	}
	return files, nil
}

func (s *Service) ListOrganizationFiles(ctx context.Context, q ListOrganizationFilesQuery) ([]ListedFile, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	allowed, err := s.accountHasOrganizationAccess(ctx, q.OrganizationID, q.Actor.AccountID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fault.Forbidden("Organization access denied")
	}
	files, err := s.repo.ListOrganizationFiles(ctx, q.OrganizationID)
	if err != nil {
		return nil, fault.Internal("List organization files failed", fault.WithCause(err))
	}
	return files, nil
}

func (s *Service) CreateMyAvatarDownload(ctx context.Context, q DownloadMyAvatarQuery) (*DownloadObjectResult, error) {
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetAccountAvatarObjectID(ctx, q.Actor.AccountID)
	if err != nil {
		return nil, fault.Internal("Resolve account avatar failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Account avatar not found", q.Actor)
}

func (s *Service) CreateMyAccountVideoDownload(ctx context.Context, q DownloadMyAccountVideoQuery) (*DownloadObjectResult, error) {
	if q.VideoID == uuid.Nil {
		return nil, fault.Validation("Account video ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetAccountVideoObjectID(ctx, q.Actor.AccountID, q.VideoID)
	if err != nil {
		return nil, fault.Internal("Resolve account video failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Account video not found", q.Actor)
}

func (s *Service) CreateOrganizationLogoDownload(ctx context.Context, q DownloadOrganizationLogoQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetOrganizationLogoObjectID(ctx, q.OrganizationID)
	if err != nil {
		return nil, fault.Internal("Resolve organization logo failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Organization logo not found", q.Actor)
}

func (s *Service) CreateOrganizationVideoDownload(ctx context.Context, q DownloadOrganizationVideoQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if q.VideoID == uuid.Nil {
		return nil, fault.Validation("Organization video ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetOrganizationVideoObjectID(ctx, q.OrganizationID, q.VideoID)
	if err != nil {
		return nil, fault.Internal("Resolve organization video failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Organization video not found", q.Actor)
}

func (s *Service) CreateCooperationPriceListDownload(ctx context.Context, q DownloadCooperationPriceListQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetCooperationPriceListObjectID(ctx, q.OrganizationID)
	if err != nil {
		return nil, fault.Internal("Resolve cooperation price list failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Cooperation price list not found", q.Actor)
}

func (s *Service) CreateOrganizationLegalDocumentDownload(ctx context.Context, q DownloadOrganizationLegalDocumentQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if q.DocumentID == uuid.Nil {
		return nil, fault.Validation("Legal document ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetOrganizationLegalDocumentObjectID(ctx, q.OrganizationID, q.DocumentID)
	if err != nil {
		return nil, fault.Internal("Resolve organization legal document failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Organization legal document not found", q.Actor)
}

func (s *Service) CreateProductImportSourceDownload(ctx context.Context, q DownloadProductImportSourceQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if q.BatchID == uuid.Nil {
		return nil, fault.Validation("Product import batch ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetProductImportSourceObjectID(ctx, q.OrganizationID, q.BatchID)
	if err != nil {
		return nil, fault.Internal("Resolve product import source failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Product import source not found", q.Actor)
}

func (s *Service) CreateProductVideoDownload(ctx context.Context, q DownloadProductVideoQuery) (*DownloadObjectResult, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, fault.Validation("Organization ID is required")
	}
	if q.ProductID == uuid.Nil {
		return nil, fault.Validation("Product ID is required")
	}
	if q.VideoID == uuid.Nil {
		return nil, fault.Validation("Product video ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	objectID, err := s.repo.GetProductVideoObjectID(ctx, q.OrganizationID, q.ProductID, q.VideoID)
	if err != nil {
		return nil, fault.Internal("Resolve product video failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Product video not found", q.Actor)
}

func (s *Service) CreateChannelAttachmentDownload(ctx context.Context, q DownloadChannelAttachmentQuery) (*DownloadObjectResult, error) {
	if q.ChannelID == uuid.Nil {
		return nil, fault.Validation("Channel ID is required")
	}
	if q.ObjectID == uuid.Nil {
		return nil, fault.Validation("Attachment object ID is required")
	}
	if !q.Actor.Authenticated {
		return nil, fault.Unauthorized("Authentication required")
	}
	exists, err := s.repo.ChannelHasAttachmentObject(ctx, q.ChannelID, q.ObjectID)
	if err != nil {
		return nil, fault.Internal("Resolve channel attachment failed", fault.WithCause(err))
	}
	if !exists {
		return nil, fault.NotFound("Channel attachment not found")
	}
	return s.CreateDownload(ctx, DownloadObjectQuery{ObjectID: q.ObjectID, Actor: q.Actor})
}

func (s *Service) CreateConferenceRecordingDownload(ctx context.Context, q DownloadConferenceRecordingQuery) (*DownloadObjectResult, error) {
	if q.ConferenceID == uuid.Nil {
		return nil, fault.Validation("Conference ID is required")
	}
	if q.RecordingID == uuid.Nil {
		return nil, fault.Validation("Recording ID is required")
	}
	if !q.Actor.Authenticated {
		return nil, fault.Unauthorized("Authentication required")
	}
	objectID, err := s.repo.GetConferenceRecordingObjectID(ctx, q.ConferenceID, q.RecordingID)
	if err != nil {
		return nil, fault.Internal("Resolve conference recording failed", fault.WithCause(err))
	}
	return s.createResolvedDownload(ctx, objectID, "Conference recording not found", q.Actor)
}

func (s *Service) ListConferenceRecordings(ctx context.Context, q ListConferenceRecordingsQuery) ([]ConferenceRecordingFile, error) {
	if q.ConferenceID == uuid.Nil {
		return nil, fault.Validation("Conference ID is required")
	}
	if !q.Actor.Authenticated {
		return nil, fault.Unauthorized("Authentication required")
	}
	allowed, err := s.canReadConference(ctx, q.ConferenceID, q.Actor)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fault.Forbidden("Conference access denied")
	}
	items, err := s.repo.ListConferenceRecordings(ctx, q.ConferenceID)
	if err != nil {
		return nil, fault.Internal("List conference recordings failed", fault.WithCause(err))
	}
	return items, nil
}

func (s *Service) canDownload(ctx context.Context, obj *StoredObject, actor authdomain.Principal) (bool, error) {
	switch {
	case actor.IsAccount():
		return s.accountCanDownload(ctx, obj, actor.AccountID)
	case actor.IsGuest():
		return s.guestCanDownload(ctx, obj, actor)
	default:
		return false, fault.Unauthorized("Authentication required")
	}
}

func (s *Service) accountCanDownload(ctx context.Context, obj *StoredObject, accountUUID uuid.UUID) (bool, error) {
	ownsAvatar, err := s.repo.AccountOwnsAvatar(ctx, accountUUID, obj.ID)
	if err != nil {
		return false, fault.Internal("Check avatar ownership failed", fault.WithCause(err))
	}
	if ownsAvatar {
		return true, nil
	}

	ownsVideo, err := s.repo.AccountOwnsVideo(ctx, accountUUID, obj.ID)
	if err != nil {
		return false, fault.Internal("Check account video ownership failed", fault.WithCause(err))
	}
	if ownsVideo {
		return true, nil
	}

	orgIDs, err := s.repo.ListRelatedOrganizationIDs(ctx, obj.ID)
	if err != nil {
		return false, fault.Internal("Resolve organization file access failed", fault.WithCause(err))
	}
	uniqueOrgIDs := uniqueUUIDs(orgIDs)
	if len(uniqueOrgIDs) > 0 {
		allowed, accessErr := s.repo.AccountHasAnyOrganizationAccess(ctx, accountUUID, uniqueOrgIDs)
		if accessErr != nil {
			return false, fault.Internal("Resolve organization membership failed", fault.WithCause(accessErr))
		}
		if allowed {
			return true, nil
		}
	}

	channelIDs, err := s.repo.ListRelatedChannelIDs(ctx, obj.ID)
	if err != nil {
		return false, fault.Internal("Resolve channel file access failed", fault.WithCause(err))
	}
	for _, channelID := range uniqueUUIDs(channelIDs) {
		if s.channels == nil {
			break
		}
		access, accessErr := s.channels.ResolveChannelAccessForAccount(ctx, channelID, accountUUID)
		if accessErr != nil {
			return false, fault.Internal("Resolve collab access failed", fault.WithCause(accessErr))
		}
		if access.Allowed && access.CanRead {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) guestCanDownload(ctx context.Context, obj *StoredObject, actor authdomain.Principal) (bool, error) {
	channelIDs, err := s.repo.ListRelatedChannelIDs(ctx, obj.ID)
	if err != nil {
		return false, fault.Internal("Resolve guest file access failed", fault.WithCause(err))
	}
	for _, channelID := range uniqueUUIDs(channelIDs) {
		if actor.ChannelID != uuid.Nil && actor.ChannelID != channelID {
			continue
		}
		if s.channels == nil {
			break
		}
		access, accessErr := s.channels.ResolveChannelAccessForGuest(ctx, channelID, actor.GuestID)
		if accessErr != nil {
			return false, fault.Internal("Resolve guest collab access failed", fault.WithCause(accessErr))
		}
		if access.Allowed && access.CanRead {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) accountHasOrganizationAccess(ctx context.Context, organizationUUID, accountUUID uuid.UUID) (bool, error) {
	if organizationUUID == uuid.Nil || accountUUID == uuid.Nil || s.memberships == nil {
		return false, nil
	}
	organizationID, err := orgdomain.OrganizationIDFromUUID(organizationUUID)
	if err != nil {
		return false, nil
	}
	accountID, err := accdomain.AccountIDFromUUID(accountUUID)
	if err != nil {
		return false, fault.Unauthorized("Authentication required")
	}
	membership, err := s.memberships.GetMemberByAccount(ctx, organizationID, accountID)
	if err != nil {
		return false, fault.Internal("Resolve organization membership failed", fault.WithCause(err))
	}
	return membership != nil && membership.IsActive() && !membership.IsRemoved(), nil
}

func (s *Service) createResolvedDownload(ctx context.Context, objectID *uuid.UUID, notFoundMessage string, actor authdomain.Principal) (*DownloadObjectResult, error) {
	if objectID == nil || *objectID == uuid.Nil {
		return nil, fault.NotFound(notFoundMessage)
	}
	return s.CreateDownload(ctx, DownloadObjectQuery{ObjectID: *objectID, Actor: actor})
}

func (s *Service) canReadConference(ctx context.Context, conferenceID uuid.UUID, actor authdomain.Principal) (bool, error) {
	channelID, err := s.repo.GetConferenceChannelID(ctx, conferenceID)
	if err != nil {
		return false, fault.Internal("Resolve conference access failed", fault.WithCause(err))
	}
	if channelID == nil || *channelID == uuid.Nil {
		return false, fault.NotFound("Conference not found")
	}
	if s.channels == nil {
		return false, nil
	}

	switch {
	case actor.IsAccount():
		access, accessErr := s.channels.ResolveChannelAccessForAccount(ctx, *channelID, actor.AccountID)
		if accessErr != nil {
			return false, fault.Internal("Resolve collab access failed", fault.WithCause(accessErr))
		}
		return access.Allowed && access.CanRead, nil
	case actor.IsGuest():
		if actor.ChannelID != uuid.Nil && actor.ChannelID != *channelID {
			return false, nil
		}
		access, accessErr := s.channels.ResolveChannelAccessForGuest(ctx, *channelID, actor.GuestID)
		if accessErr != nil {
			return false, fault.Internal("Resolve guest collab access failed", fault.WithCause(accessErr))
		}
		return access.Allowed && access.CanRead, nil
	default:
		return false, fault.Unauthorized("Authentication required")
	}
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
