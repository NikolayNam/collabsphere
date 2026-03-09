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
	object       *StoredObject
	avatarOwned  bool
	orgIDs       []uuid.UUID
	channelIDs   []uuid.UUID
	accountFiles []ListedFile
	orgFiles     []ListedFile
}

func (r repoStub) GetObjectByID(ctx context.Context, objectID uuid.UUID) (*StoredObject, error) {
	return r.object, nil
}
func (r repoStub) AccountOwnsAvatar(ctx context.Context, accountID, objectID uuid.UUID) (bool, error) {
	return r.avatarOwned, nil
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
		SourceType: "account_avatar",
	}}
	svc := New(repoStub{accountFiles: expected}, nil, nil, nil)

	result, err := svc.ListMyFiles(context.Background(), ListMyFilesQuery{Actor: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())})
	if err != nil {
		t.Fatalf("ListMyFiles() error = %v", err)
	}
	if len(result) != 1 || result[0].SourceType != "account_avatar" {
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
	expected := []ListedFile{{
		ObjectID:       uuid.New(),
		OrganizationID: &orgUUID,
		FileName:       "logo.png",
		SizeBytes:      128,
		CreatedAt:      now,
		SourceType:     "organization_logo",
		SourceID:       &orgUUID,
	}}
	svc := New(repoStub{orgFiles: expected}, membershipsStub{member: member}, nil, nil)

	result, err := svc.ListOrganizationFiles(context.Background(), ListOrganizationFilesQuery{
		OrganizationID: orgUUID,
		Actor:          authdomain.NewAccountPrincipal(accountID.UUID(), uuid.New()),
	})
	if err != nil {
		t.Fatalf("ListOrganizationFiles() error = %v", err)
	}
	if len(result) != 1 || result[0].SourceType != "organization_logo" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
