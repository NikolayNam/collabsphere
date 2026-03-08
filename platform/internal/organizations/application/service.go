package application

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization"
	create_with_owner "github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization_with_owner"
	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/get_organization_by_id"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

var (
	ErrValidation = apperrors.ErrValidation
)

type CreateOrganizationCmd = create_organization.Command
type GetOrganizationByIdQuery = get_organization_by_id.Query

type UpdateOrganizationProfileCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	Name           *string
	Slug           *string
	LogoObjectID   *uuid.UUID
	ClearLogo      bool
	Description    *string
	Website        *string
	PrimaryEmail   *string
	Phone          *string
	Address        *string
	Industry       *string
}

type CreateOrganizationLogoUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type CreateOrganizationLogoUploadResult struct {
	ObjectID  uuid.UUID
	Bucket    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
	FileName  string
	SizeBytes int64
}

type Service struct {
	create      *create_organization.Handler
	getById     *get_organization_by_id.Handler
	repo        ports.OrganizationRepository
	memberships memberPorts.MembershipRepository
	clock       ports.Clock
	storage     ports.ObjectStorage
	bucket      string
}

func New(repo ports.OrganizationRepository, membershipRepo memberPorts.MembershipRepository, categoryProvisioner ports.ProductCategoryProvisioner, txm sharedtx.Manager, clock ports.Clock, storage ports.ObjectStorage, bucket string) *Service {
	creator := create_with_owner.New(txm, repo, membershipRepo, categoryProvisioner)

	return &Service{
		create:      create_organization.NewHandler(creator, clock),
		getById:     get_organization_by_id.NewHandler(repo),
		repo:        repo,
		memberships: membershipRepo,
		clock:       clock,
		storage:     storage,
		bucket:      strings.TrimSpace(bucket),
	}
}

func (s *Service) CreateOrganization(ctx context.Context, cmd CreateOrganizationCmd) (*domain.Organization, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetOrganizationById(ctx context.Context, q GetOrganizationByIdQuery) (*domain.Organization, error) {
	return s.getById.Handle(ctx, q)
}

func (s *Service) UpdateOrganizationProfile(ctx context.Context, cmd UpdateOrganizationProfileCmd) (*domain.Organization, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.ClearLogo && cmd.LogoObjectID != nil {
		return nil, apperrors.InvalidInput("clearLogo and logoObjectId cannot be used together")
	}
	updated, err := s.repo.UpdateProfile(ctx, cmd.OrganizationID, domain.OrganizationProfilePatch{
		Name:         cmd.Name,
		Slug:         cmd.Slug,
		LogoObjectID: cmd.LogoObjectID,
		ClearLogo:    cmd.ClearLogo,
		Description:  cmd.Description,
		Website:      cmd.Website,
		PrimaryEmail: cmd.PrimaryEmail,
		Phone:        cmd.Phone,
		Address:      cmd.Address,
		Industry:     cmd.Industry,
		UpdatedAt:    s.clock.Now(),
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperrors.OrganizationNotFound()
	}
	return updated, nil
}

func (s *Service) CreateOrganizationLogoUpload(ctx context.Context, cmd CreateOrganizationLogoUploadCmd) (*CreateOrganizationLogoUploadResult, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if s.storage == nil || s.bucket == "" {
		return nil, fault.Unavailable("Logo upload is unavailable")
	}
	fileName := strings.TrimSpace(cmd.FileName)
	if fileName == "" {
		return nil, apperrors.InvalidInput("fileName is required")
	}

	sizeBytes := int64(0)
	if cmd.SizeBytes != nil {
		if *cmd.SizeBytes < 0 {
			return nil, apperrors.InvalidInput("sizeBytes must be non-negative")
		}
		sizeBytes = *cmd.SizeBytes
	}

	now := s.clock.Now()
	objectID := uuid.New()
	orgUUID := cmd.OrganizationID.UUID()
	objectKey := buildOrganizationLogoObjectKey(orgUUID, objectID, fileName, now)
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: &orgUUID,
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       sanitizeFileName(fileName, "logo.bin"),
		ContentType:    normalizeOptional(cmd.ContentType),
		SizeBytes:      sizeBytes,
		ChecksumSHA256: normalizeOptional(cmd.ChecksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create logo object failed", fault.WithCause(err))
	}
	uploadURL, expiresAt, err := s.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign logo upload failed", fault.WithCause(err))
	}
	return &CreateOrganizationLogoUploadResult{
		ObjectID:  object.ID,
		Bucket:    object.Bucket,
		ObjectKey: object.ObjectKey,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
		FileName:  object.FileName,
		SizeBytes: object.SizeBytes,
	}, nil
}

func (s *Service) requireOrganizationAccess(ctx context.Context, organizationID domain.OrganizationID, actorAccountID uuid.UUID, requireOwner bool) error {
	if organizationID.IsZero() {
		return apperrors.InvalidInput("Organization is required")
	}
	if actorAccountID == uuid.Nil {
		return fault.Unauthorized("Authentication required")
	}

	organization, err := s.repo.GetByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if organization == nil {
		return apperrors.OrganizationNotFound()
	}

	members, err := s.memberships.ListMembers(ctx, organizationID)
	if err != nil {
		return fault.Internal("List organization members failed", fault.WithCause(err))
	}
	for _, member := range members {
		if member.AccountID != actorAccountID || !member.IsActive {
			continue
		}
		if requireOwner && member.Role != string(memberdomain.MembershipRoleOwner) {
			return fault.Forbidden("Only organization owners can manage organization profile")
		}
		return nil
	}
	return fault.Forbidden("Organization access denied")
}

func buildOrganizationLogoObjectKey(organizationID, objectID uuid.UUID, fileName string, now time.Time) string {
	return strings.Join([]string{
		"organizations",
		"logos",
		organizationID.String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		sanitizeFileName(fileName, "logo.bin"),
	}, "/")
}

func sanitizeFileName(fileName, fallback string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return fallback
	}

	var b strings.Builder
	b.Grow(len(base))
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
		return fallback
	}
	return out
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
