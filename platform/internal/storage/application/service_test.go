package application

import (
	"context"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type storageStub struct {
	url string
	exp time.Time
}

func (s storageStub) PresignGetObject(ctx context.Context, bucket, objectKey string) (string, time.Time, error) {
	return s.url, s.exp, nil
}

type repoStub struct {
	object                    *StoredObject
	avatarObjectID            *uuid.UUID
	accountVideoObjectID      *uuid.UUID
	organizationLogoObjectID  *uuid.UUID
	organizationVideoObjectID *uuid.UUID
	priceListObjectID         *uuid.UUID
	legalDocumentObjectID     *uuid.UUID
	productImportSourceObject *uuid.UUID
	productVideoObject        *uuid.UUID
	conferenceChannelID       *uuid.UUID
	conferenceRecordingObject *uuid.UUID
	conferenceRecordings      []ConferenceRecordingFile
	channelHasAttachment      bool
	avatarOwned               bool
	accountVideoOwned         bool
	orgIDs                    []uuid.UUID
	channelIDs                []uuid.UUID
	accountFiles              []ListedFile
	orgFiles                  []ListedFile
}

func (r repoStub) GetObjectByID(ctx context.Context, objectID uuid.UUID) (*StoredObject, error) {
	return r.object, nil
}
func (r repoStub) GetAccountAvatarObjectID(ctx context.Context, accountID uuid.UUID) (*uuid.UUID, error) {
	return r.avatarObjectID, nil
}
func (r repoStub) GetAccountVideoObjectID(ctx context.Context, accountID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.accountVideoObjectID, nil
}
func (r repoStub) GetOrganizationLogoObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error) {
	return r.organizationLogoObjectID, nil
}
func (r repoStub) GetOrganizationVideoObjectID(ctx context.Context, organizationID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.organizationVideoObjectID, nil
}
func (r repoStub) GetCooperationPriceListObjectID(ctx context.Context, organizationID uuid.UUID) (*uuid.UUID, error) {
	return r.priceListObjectID, nil
}
func (r repoStub) GetOrganizationLegalDocumentObjectID(ctx context.Context, organizationID, documentID uuid.UUID) (*uuid.UUID, error) {
	return r.legalDocumentObjectID, nil
}
func (r repoStub) GetProductImportSourceObjectID(ctx context.Context, organizationID, batchID uuid.UUID) (*uuid.UUID, error) {
	return r.productImportSourceObject, nil
}
func (r repoStub) GetProductVideoObjectID(ctx context.Context, organizationID, productID, videoID uuid.UUID) (*uuid.UUID, error) {
	return r.productVideoObject, nil
}
func (r repoStub) GetConferenceChannelID(ctx context.Context, conferenceID uuid.UUID) (*uuid.UUID, error) {
	return r.conferenceChannelID, nil
}
func (r repoStub) GetConferenceRecordingObjectID(ctx context.Context, conferenceID, recordingID uuid.UUID) (*uuid.UUID, error) {
	return r.conferenceRecordingObject, nil
}
func (r repoStub) ListConferenceRecordings(ctx context.Context, conferenceID uuid.UUID) ([]ConferenceRecordingFile, error) {
	return r.conferenceRecordings, nil
}
func (r repoStub) ChannelHasAttachmentObject(ctx context.Context, channelID, objectID uuid.UUID) (bool, error) {
	return r.channelHasAttachment, nil
}
func (r repoStub) AccountOwnsAvatar(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	return r.avatarOwned, nil
}
func (r repoStub) AccountOwnsVideo(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	return r.accountVideoOwned, nil
}
func (r repoStub) ListRelatedOrganizationIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error) {
	return r.orgIDs, nil
}
func (r repoStub) ListRelatedChannelIDs(ctx context.Context, objectID uuid.UUID) ([]uuid.UUID, error) {
	return r.channelIDs, nil
}
func (r repoStub) ListAccountFiles(ctx context.Context, accountID uuid.UUID) ([]ListedFile, error) {
	return r.accountFiles, nil
}
func (r repoStub) ListOrganizationFiles(ctx context.Context, organizationID uuid.UUID) ([]ListedFile, error) {
	return r.orgFiles, nil
}

type membershipsStub struct {
	member *memberdomain.Membership
}

func (m membershipsStub) AddMember(ctx context.Context, orgID orgdomain.OrganizationID, membership *memberdomain.Membership) error {
	panic("unexpected call")
}
func (m membershipsStub) SaveMember(ctx context.Context, orgID orgdomain.OrganizationID, membership *memberdomain.Membership) error {
	panic("unexpected call")
}
func (m membershipsStub) GetMemberByAccount(ctx context.Context, orgID orgdomain.OrganizationID, accountID accdomain.AccountID) (*memberdomain.Membership, error) {
	return m.member, nil
}
func (m membershipsStub) GetMemberByID(ctx context.Context, orgID orgdomain.OrganizationID, membershipID uuid.UUID) (*memberdomain.Membership, error) {
	panic("unexpected call")
}
func (m membershipsStub) CountActiveMembersByRole(ctx context.Context, orgID orgdomain.OrganizationID, role memberdomain.MembershipRole) (int64, error) {
	panic("unexpected call")
}
func (m membershipsStub) ListMembers(ctx context.Context, orgID orgdomain.OrganizationID) ([]memberdomain.MemberView, error) {
	panic("unexpected call")
}
func (m membershipsStub) CreateInvitation(ctx context.Context, invitation *memberdomain.OrganizationInvitation) error {
	panic("unexpected call")
}
func (m membershipsStub) SaveInvitation(ctx context.Context, invitation *memberdomain.OrganizationInvitation) error {
	panic("unexpected call")
}
func (m membershipsStub) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*memberdomain.OrganizationInvitation, error) {
	panic("unexpected call")
}
func (m membershipsStub) ListInvitations(ctx context.Context, orgID orgdomain.OrganizationID) ([]memberdomain.OrganizationInvitation, error) {
	panic("unexpected call")
}
func (m membershipsStub) RevokeExpiredPendingInvitations(ctx context.Context, orgID orgdomain.OrganizationID, email accdomain.Email, actorAccountID uuid.UUID, now time.Time) error {
	panic("unexpected call")
}

type channelAccessStub struct {
	accountAccess collabdomain.Access
	guestAccess   collabdomain.Access
}

func (c channelAccessStub) ResolveChannelAccessForAccount(ctx context.Context, channelID, accountID uuid.UUID) (collabdomain.Access, error) {
	return c.accountAccess, nil
}
func (c channelAccessStub) ResolveChannelAccessForGuest(ctx context.Context, channelID, guestID uuid.UUID) (collabdomain.Access, error) {
	return c.guestAccess, nil
}

func TestCreateDownloadAllowsOwnAvatar(t *testing.T) {
	objectID := uuid.New()
	now := time.Now().UTC()
	svc := New(
		repoStub{object: &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "avatars/me.png", FileName: "me.png", SizeBytes: 42, CreatedAt: now}, avatarOwned: true},
		membershipsStub{},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateDownload(context.Background(), DownloadObjectQuery{ObjectID: objectID, Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	if result.DownloadURL != "http://example.com/download" {
		t.Fatalf("unexpected download URL %q", result.DownloadURL)
	}
}

func TestCreateDownloadAllowsOwnAccountVideo(t *testing.T) {
	objectID := uuid.New()
	now := time.Now().UTC()
	svc := New(
		repoStub{object: &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "accounts/videos/me.mp4", FileName: "me.mp4", SizeBytes: 42, CreatedAt: now}, accountVideoOwned: true},
		membershipsStub{},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateDownload(context.Background(), DownloadObjectQuery{ObjectID: objectID, Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateDownloadAllowsOrganizationMember(t *testing.T) {
	objectID := uuid.New()
	organizationUUID := uuid.New()
	now := time.Now().UTC()
	organizationID, _ := orgdomain.OrganizationIDFromUUID(organizationUUID)
	accountID, _ := accdomain.AccountIDFromUUID(uuid.New())
	membership, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: organizationID,
		AccountID:      accountID,
		Role:           memberdomain.MembershipRoleMember,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	actor := authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New())
	svc := New(
		repoStub{object: &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "org/file.pdf", FileName: "file.pdf", SizeBytes: 42, CreatedAt: now}, orgIDs: []uuid.UUID{organizationUUID}},
		membershipsStub{member: membership},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateDownload(context.Background(), DownloadObjectQuery{ObjectID: objectID, Actor: actor})
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateDownloadAllowsCollabGuestAttachment(t *testing.T) {
	objectID := uuid.New()
	channelID := uuid.New()
	now := time.Now().UTC()
	svc := New(
		repoStub{object: &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "collab/file.bin", FileName: "file.bin", SizeBytes: 42, CreatedAt: now}, channelIDs: []uuid.UUID{channelID}},
		nil,
		channelAccessStub{guestAccess: collabdomain.Access{Allowed: true, CanRead: true}},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateDownload(context.Background(), DownloadObjectQuery{ObjectID: objectID, Actor: authdomain.NewGuestPrincipal(uuid.New(), uuid.New(), channelID)})
	if err != nil {
		t.Fatalf("CreateDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateDownloadRejectsUnauthorizedObject(t *testing.T) {
	objectID := uuid.New()
	now := time.Now().UTC()
	svc := New(
		repoStub{object: &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "private/file.bin", FileName: "file.bin", SizeBytes: 42, CreatedAt: now}},
		nil,
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	_, err := svc.CreateDownload(context.Background(), DownloadObjectQuery{ObjectID: objectID, Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMyFilesRequiresAccount(t *testing.T) {
	svc := New(repoStub{}, nil, nil, nil)

	_, err := svc.ListMyFiles(context.Background(), ListMyFilesQuery{Actor: authdomain.AnonymousPrincipal()})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListMyFilesReturnsAccountFiles(t *testing.T) {
	now := time.Now().UTC()
	expected := []ListedFile{{
		ObjectID:   uuid.New(),
		FileName:   "avatar.png",
		SizeBytes:  42,
		CreatedAt:  now,
		SourceType: "account_video",
	}}
	svc := New(repoStub{accountFiles: expected}, nil, nil, nil)

	result, err := svc.ListMyFiles(context.Background(), ListMyFilesQuery{Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err != nil {
		t.Fatalf("ListMyFiles() error = %v", err)
	}
	if len(result) != 1 || result[0].SourceType != "account_video" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestListOrganizationFilesRequiresMembership(t *testing.T) {
	orgUUID := uuid.New()
	svc := New(repoStub{}, membershipsStub{}, nil, nil)

	_, err := svc.ListOrganizationFiles(context.Background(), ListOrganizationFilesQuery{
		OrganizationID: orgUUID,
		Actor:          authdomain.NewAccountPrincipal(uuid.New(), uuid.New()),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListOrganizationFilesReturnsOrganizationFiles(t *testing.T) {
	now := time.Now().UTC()
	orgUUID := uuid.New()
	orgID, _ := orgdomain.OrganizationIDFromUUID(orgUUID)
	accountID, _ := accdomain.AccountIDFromUUID(uuid.New())
	member, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: orgID,
		AccountID:      accountID,
		Role:           memberdomain.MembershipRoleViewer,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	productID := uuid.New()
	expected := []ListedFile{{
		ObjectID:       uuid.New(),
		OrganizationID: &orgUUID,
		FileName:       "presentation.mp4",
		SizeBytes:      1024,
		CreatedAt:      now,
		SourceType:     "product_video",
		SourceID:       &productID,
	}}
	svc := New(repoStub{orgFiles: expected}, membershipsStub{member: member}, nil, nil)

	result, err := svc.ListOrganizationFiles(context.Background(), ListOrganizationFilesQuery{
		OrganizationID: orgUUID,
		Actor:          authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New()),
	})
	if err != nil {
		t.Fatalf("ListOrganizationFiles() error = %v", err)
	}
	if len(result) != 1 || result[0].SourceType != "product_video" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCreateMyAvatarDownloadResolvesAvatarObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	accountID := uuid.New()
	svc := New(
		repoStub{
			avatarObjectID: &objectID,
			object:         &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "avatars/me.png", FileName: "me.png", SizeBytes: 42, CreatedAt: now},
			avatarOwned:    true,
		},
		membershipsStub{},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateMyAvatarDownload(context.Background(), DownloadMyAvatarQuery{Actor: authdomain.NewAccountPrincipal(accountID, uuid.New())})
	if err != nil {
		t.Fatalf("CreateMyAvatarDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateMyAccountVideoDownloadResolvesVideoObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	videoID := uuid.New()
	accountID := uuid.New()
	svc := New(
		repoStub{
			accountVideoObjectID: &objectID,
			object:               &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "accounts/videos/me.mp4", FileName: "me.mp4", SizeBytes: 42, CreatedAt: now},
			accountVideoOwned:    true,
		},
		membershipsStub{},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateMyAccountVideoDownload(context.Background(), DownloadMyAccountVideoQuery{VideoID: videoID, Actor: authdomain.NewAccountPrincipal(accountID, uuid.New())})
	if err != nil {
		t.Fatalf("CreateMyAccountVideoDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateOrganizationLogoDownloadResolvesLogoObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	organizationUUID := uuid.New()
	organizationID, _ := orgdomain.OrganizationIDFromUUID(organizationUUID)
	accountID, _ := accdomain.AccountIDFromUUID(uuid.New())
	membership, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: organizationID,
		AccountID:      accountID,
		Role:           memberdomain.MembershipRoleViewer,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	svc := New(
		repoStub{
			organizationLogoObjectID: &objectID,
			object:                   &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "org/logo.png", FileName: "logo.png", SizeBytes: 64, CreatedAt: now},
			orgIDs:                   []uuid.UUID{organizationUUID},
		},
		membershipsStub{member: membership},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateOrganizationLogoDownload(context.Background(), DownloadOrganizationLogoQuery{OrganizationID: organizationUUID, Actor: authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New())})
	if err != nil {
		t.Fatalf("CreateOrganizationLogoDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateOrganizationVideoDownloadResolvesVideoObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	videoID := uuid.New()
	organizationUUID := uuid.New()
	organizationID, _ := orgdomain.OrganizationIDFromUUID(organizationUUID)
	accountID, _ := accdomain.AccountIDFromUUID(uuid.New())
	membership, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: organizationID,
		AccountID:      accountID,
		Role:           memberdomain.MembershipRoleMember,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	svc := New(
		repoStub{
			organizationVideoObjectID: &objectID,
			object:                    &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "org/videos/brand.mp4", FileName: "brand.mp4", SizeBytes: 64, CreatedAt: now},
			orgIDs:                    []uuid.UUID{organizationUUID},
		},
		membershipsStub{member: membership},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateOrganizationVideoDownload(context.Background(), DownloadOrganizationVideoQuery{OrganizationID: organizationUUID, VideoID: videoID, Actor: authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New())})
	if err != nil {
		t.Fatalf("CreateOrganizationVideoDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateChannelAttachmentDownloadRequiresAttachedObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	channelID := uuid.New()
	svc := New(
		repoStub{
			channelHasAttachment: true,
			object:               &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "collab/file.bin", FileName: "file.bin", SizeBytes: 42, CreatedAt: now},
			channelIDs:           []uuid.UUID{channelID},
		},
		nil,
		channelAccessStub{accountAccess: collabdomain.Access{Allowed: true, CanRead: true}},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateChannelAttachmentDownload(context.Background(), DownloadChannelAttachmentQuery{ChannelID: channelID, ObjectID: objectID, Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err != nil {
		t.Fatalf("CreateChannelAttachmentDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestCreateConferenceRecordingDownloadResolvesRecording(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	channelID := uuid.New()
	conferenceID := uuid.New()
	recordingID := uuid.New()
	svc := New(
		repoStub{
			conferenceRecordingObject: &objectID,
			object:                    &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "collab/recording.mp4", FileName: "recording.mp4", SizeBytes: 42, CreatedAt: now},
			channelIDs:                []uuid.UUID{channelID},
		},
		nil,
		channelAccessStub{accountAccess: collabdomain.Access{Allowed: true, CanRead: true}},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateConferenceRecordingDownload(context.Background(), DownloadConferenceRecordingQuery{
		ConferenceID: conferenceID,
		RecordingID:  recordingID,
		Actor:        authdomain.NewAccountPrincipal(uuid.New(), uuid.New()),
	})
	if err != nil {
		t.Fatalf("CreateConferenceRecordingDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}

func TestListConferenceRecordingsReturnsItems(t *testing.T) {
	now := time.Now().UTC()
	conferenceID := uuid.New()
	channelID := uuid.New()
	svc := New(
		repoStub{
			conferenceChannelID: &channelID,
			conferenceRecordings: []ConferenceRecordingFile{{
				RecordingID:  uuid.New(),
				ConferenceID: conferenceID,
				ObjectID:     uuid.New(),
				FileName:     "recording.mp4",
				SizeBytes:    1024,
				CreatedAt:    now,
			}},
		},
		nil,
		channelAccessStub{accountAccess: collabdomain.Access{Allowed: true, CanRead: true}},
		nil,
	)

	items, err := svc.ListConferenceRecordings(context.Background(), ListConferenceRecordingsQuery{
		ConferenceID: conferenceID,
		Actor:        authdomain.NewAccountPrincipal(uuid.New(), uuid.New()),
	})
	if err != nil {
		t.Fatalf("ListConferenceRecordings() error = %v", err)
	}
	if len(items) != 1 || items[0].ConferenceID != conferenceID {
		t.Fatalf("unexpected result: %#v", items)
	}
}

func TestCreateProductVideoDownloadResolvesObject(t *testing.T) {
	now := time.Now().UTC()
	objectID := uuid.New()
	organizationUUID := uuid.New()
	productID := uuid.New()
	videoID := uuid.New()
	organizationID, _ := orgdomain.OrganizationIDFromUUID(organizationUUID)
	accountID, _ := accdomain.AccountIDFromUUID(uuid.New())
	membership, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: organizationID,
		AccountID:      accountID,
		Role:           memberdomain.MembershipRoleMember,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	svc := New(
		repoStub{
			productVideoObject: &objectID,
			object:             &StoredObject{ID: objectID, Bucket: "collabsphere", ObjectKey: "catalog/products/video.mp4", FileName: "video.mp4", SizeBytes: 128, CreatedAt: now},
			orgIDs:             []uuid.UUID{organizationUUID},
		},
		membershipsStub{member: membership},
		channelAccessStub{},
		storageStub{url: "http://example.com/download", exp: now.Add(5 * time.Minute)},
	)

	result, err := svc.CreateProductVideoDownload(context.Background(), DownloadProductVideoQuery{
		OrganizationID: organizationUUID,
		ProductID:      productID,
		VideoID:        videoID,
		Actor:          authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New()),
	})
	if err != nil {
		t.Fatalf("CreateProductVideoDownload() error = %v", err)
	}
	if result.ObjectID != objectID {
		t.Fatalf("unexpected object id %s", result.ObjectID)
	}
}
