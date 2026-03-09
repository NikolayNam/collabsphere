package application

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/accounts/application/create_account"
	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/get_account_by_email"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/get_account_by_id"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
)

var (
	ErrValidation = apperrors.ErrValidation
	ErrNotFound   = apperrors.ErrNotFound
)

type CreateAccountCmd = create_account.Command
type GetAccountByIdQuery = get_account_by_id.Query
type GetAccountByEmailQuery = get_account_by_email.Query

type UpdateMyProfileCmd struct {
	AccountID      domain.AccountID
	DisplayName    *string
	AvatarObjectID *uuid.UUID
	ClearAvatar    bool
	Bio            *string
	Phone          *string
	Locale         *string
	Timezone       *string
	Website        *string
}

type CreateAvatarUploadCmd struct {
	AccountID      domain.AccountID
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type UploadAvatarCmd struct {
	AccountID   domain.AccountID
	FileName    string
	ContentType string
	SizeBytes   int64
	Body        io.Reader
}

type CreateAvatarUploadResult struct {
	ObjectID  uuid.UUID
	Bucket    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
	FileName  string
	SizeBytes int64
}

type Service struct {
	create     *create_account.Handler
	getById    *get_account_by_id.Handler
	getByEmail *get_account_by_email.Handler
	repo       ports.AccountRepository
	clock      ports.Clock
	storage    ports.ObjectStorage
	bucket     string
}

func New(repo ports.AccountRepository, hasher ports.PasswordHasher, clock ports.Clock, storage ports.ObjectStorage, bucket string) *Service {
	return &Service{
		create:     create_account.NewHandler(repo, hasher, clock),
		getById:    get_account_by_id.NewHandler(repo),
		getByEmail: get_account_by_email.NewHandler(repo),
		repo:       repo,
		clock:      clock,
		storage:    storage,
		bucket:     strings.TrimSpace(bucket),
	}
}

func (s *Service) CreateAccount(ctx context.Context, cmd CreateAccountCmd) (*domain.Account, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetAccountById(ctx context.Context, q GetAccountByIdQuery) (*domain.Account, error) {
	return s.getById.Handle(ctx, q)
}

func (s *Service) GetAccountByEmail(ctx context.Context, q GetAccountByEmailQuery) (*domain.Account, error) {
	return s.getByEmail.Handle(ctx, q)
}

func (s *Service) GetMyProfile(ctx context.Context, accountID domain.AccountID) (*domain.Account, error) {
	if accountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, apperrors.AccountNotFound()
	}
	return acc, nil
}

func (s *Service) UpdateMyProfile(ctx context.Context, cmd UpdateMyProfileCmd) (*domain.Account, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if cmd.ClearAvatar && cmd.AvatarObjectID != nil {
		return nil, apperrors.InvalidInput("clearAvatar and avatarObjectId cannot be used together")
	}
	updated, err := s.repo.UpdateProfile(ctx, cmd.AccountID, domain.AccountProfilePatch{
		DisplayName:    cmd.DisplayName,
		AvatarObjectID: cmd.AvatarObjectID,
		ClearAvatar:    cmd.ClearAvatar,
		Bio:            cmd.Bio,
		Phone:          cmd.Phone,
		Locale:         cmd.Locale,
		Timezone:       cmd.Timezone,
		Website:        cmd.Website,
		UpdatedAt:      s.clock.Now(),
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperrors.AccountNotFound()
	}
	return updated, nil
}

func (s *Service) CreateAvatarUpload(ctx context.Context, cmd CreateAvatarUploadCmd) (*CreateAvatarUploadResult, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if s.storage == nil || s.bucket == "" {
		return nil, fault.Unavailable("Avatar upload is unavailable")
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
	objectKey := buildAccountAvatarObjectKey(cmd.AccountID.UUID(), objectID, fileName, now)
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: nil,
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       sanitizeFileName(fileName, "avatar.bin"),
		ContentType:    normalizeOptional(cmd.ContentType),
		SizeBytes:      sizeBytes,
		ChecksumSHA256: normalizeOptional(cmd.ChecksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create avatar object failed", fault.WithCause(err))
	}
	uploadURL, expiresAt, err := s.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign avatar upload failed", fault.WithCause(err))
	}
	return &CreateAvatarUploadResult{
		ObjectID:  object.ID,
		Bucket:    object.Bucket,
		ObjectKey: object.ObjectKey,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
		FileName:  object.FileName,
		SizeBytes: object.SizeBytes,
	}, nil
}

func (s *Service) UploadAvatar(ctx context.Context, cmd UploadAvatarCmd) (*domain.Account, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if s.storage == nil || s.bucket == "" {
		return nil, fault.Unavailable("Avatar upload is unavailable")
	}
	if cmd.Body == nil {
		return nil, apperrors.InvalidInput("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, apperrors.InvalidInput("file size must be non-negative")
	}

	account, err := s.repo.GetByID(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, apperrors.AccountNotFound()
	}

	fileName := sanitizeFileName(strings.TrimSpace(cmd.FileName), "avatar.bin")
	contentType := strings.TrimSpace(cmd.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	now := s.clock.Now()
	objectID := uuid.New()
	objectKey := buildAccountAvatarObjectKey(cmd.AccountID.UUID(), objectID, fileName, now)
	objectContentType := &contentType
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: nil,
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    objectContentType,
		SizeBytes:      cmd.SizeBytes,
		ChecksumSHA256: nil,
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create avatar object failed", fault.WithCause(err))
	}
	if err := s.storage.PutObject(ctx, object.Bucket, object.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload avatar failed", fault.WithCause(err))
	}

	updated, err := s.repo.UpdateProfile(ctx, cmd.AccountID, domain.AccountProfilePatch{
		AvatarObjectID: &objectID,
		UpdatedAt:      now,
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperrors.AccountNotFound()
	}
	return updated, nil
}

func buildAccountAvatarObjectKey(accountID, objectID uuid.UUID, fileName string, now time.Time) string {
	return strings.Join([]string{
		"accounts",
		"avatars",
		accountID.String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		sanitizeFileName(fileName, "avatar.bin"),
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
