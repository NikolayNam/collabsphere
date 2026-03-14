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
	uploadports "github.com/NikolayNam/collabsphere/internal/uploads/application/ports"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	sharedkyc "github.com/NikolayNam/collabsphere/shared/kyc"
)

var (
	ErrValidation = apperrors.ErrValidation
	ErrNotFound   = apperrors.ErrNotFound
)

type CreateAccountCmd = create_account.Command
type GetAccountByIdQuery = get_account_by_id.Query
type GetAccountByEmailQuery = get_account_by_email.Query

type AccountVideoView struct {
	ID          uuid.UUID
	ObjectID    uuid.UUID
	FileName    string
	ContentType *string
	SizeBytes   int64
	CreatedAt   time.Time
	SortOrder   int64
}

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

type UploadMyVideoCmd struct {
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

type AccountKYCProfileView struct {
	AccountID        uuid.UUID
	Status           string
	LegalName        *string
	CountryCode      *string
	DocumentNumber   *string
	ResidenceAddress *string
	ReviewNote       *string
	ReviewerAccount  *uuid.UUID
	SubmittedAt      *time.Time
	ReviewedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type AccountKYCDocumentView struct {
	ID              uuid.UUID
	AccountID       uuid.UUID
	ObjectID        uuid.UUID
	DocumentType    string
	Title           string
	Status          string
	ReviewNote      *string
	ReviewerAccount *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	ReviewedAt      *time.Time
}

type UpdateMyKYCProfileCmd struct {
	AccountID        domain.AccountID
	Status           *string
	LegalName        *string
	CountryCode      *string
	DocumentNumber   *string
	ResidenceAddress *string
}

type CreateMyKYCDocumentUploadCmd struct {
	AccountID      domain.AccountID
	DocumentType   string
	Title          *string
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type CreateMyKYCDocumentUploadResult struct {
	UploadID     uuid.UUID
	ObjectID     uuid.UUID
	Bucket       string
	ObjectKey    string
	UploadURL    string
	ExpiresAt    time.Time
	FileName     string
	SizeBytes    int64
	DocumentType string
	Title        string
}

type CompleteMyKYCDocumentUploadCmd struct {
	AccountID domain.AccountID
	UploadID  uuid.UUID
}

type Service struct {
	create     *create_account.Handler
	getById    *get_account_by_id.Handler
	getByEmail *get_account_by_email.Handler
	repo       ports.AccountRepository
	clock      ports.Clock
	storage    ports.ObjectStorage
	bucket     string
	uploads    uploadports.Repository
}

func New(repo ports.AccountRepository, hasher ports.PasswordHasher, clock ports.Clock, storage ports.ObjectStorage, bucket string, uploads uploadports.Repository) *Service {
	return &Service{
		create:     create_account.NewHandler(repo, hasher, clock),
		getById:    get_account_by_id.NewHandler(repo),
		getByEmail: get_account_by_email.NewHandler(repo),
		repo:       repo,
		clock:      clock,
		storage:    storage,
		bucket:     strings.TrimSpace(bucket),
		uploads:    uploads,
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

func (s *Service) UploadMyVideo(ctx context.Context, cmd UploadMyVideoCmd) (*ports.AccountVideoRecord, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if s.storage == nil || s.bucket == "" {
		return nil, fault.Unavailable("Account video upload is unavailable")
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
	fileName := sanitizeFileName(strings.TrimSpace(cmd.FileName), "video.mp4")
	contentType, err := normalizeVideoContentType(fileName, cmd.ContentType)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	objectID := uuid.New()
	objectKey := buildAccountVideoObjectKey(cmd.AccountID.UUID(), objectID, fileName, now)
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: nil,
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    &contentType,
		SizeBytes:      cmd.SizeBytes,
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create account video object failed", fault.WithCause(err))
	}
	if err := s.storage.PutObject(ctx, object.Bucket, object.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload account video failed", fault.WithCause(err))
	}
	video, err := s.repo.CreateAccountVideo(ctx, cmd.AccountID.UUID(), objectID, now)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (s *Service) ListMyVideos(ctx context.Context, accountID domain.AccountID) ([]AccountVideoView, error) {
	if accountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	items, err := s.repo.ListAccountVideos(ctx, accountID.UUID())
	if err != nil {
		return nil, err
	}
	out := make([]AccountVideoView, 0, len(items))
	for _, item := range items {
		out = append(out, AccountVideoView{
			ID:          item.ID,
			ObjectID:    item.ObjectID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			CreatedAt:   item.CreatedAt,
			SortOrder:   item.SortOrder,
		})
	}
	return out, nil
}

func (s *Service) ListMyVideoObjectIDs(ctx context.Context, accountID domain.AccountID) ([]uuid.UUID, error) {
	if accountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	return s.repo.ListAccountVideoObjectIDs(ctx, accountID.UUID())
}

func (s *Service) GetMyKYCProfile(ctx context.Context, accountID domain.AccountID) (*AccountKYCProfileView, []AccountKYCDocumentView, error) {
	if accountID.IsZero() {
		return nil, nil, apperrors.InvalidInput("Account ID is required")
	}
	profile, err := s.repo.GetAccountKYCProfile(ctx, accountID.UUID())
	if err != nil {
		return nil, nil, err
	}
	documents, err := s.repo.ListAccountKYCDocuments(ctx, accountID.UUID())
	if err != nil {
		return nil, nil, err
	}
	documentViews := make([]AccountKYCDocumentView, 0, len(documents))
	for _, item := range documents {
		documentViews = append(documentViews, AccountKYCDocumentView{
			ID:              item.ID,
			AccountID:       item.AccountID,
			ObjectID:        item.ObjectID,
			DocumentType:    item.DocumentType,
			Title:           item.Title,
			Status:          item.Status,
			ReviewNote:      item.ReviewNote,
			ReviewerAccount: item.ReviewerAccount,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
			ReviewedAt:      item.ReviewedAt,
		})
	}
	if profile == nil {
		now := s.clock.Now()
		return &AccountKYCProfileView{
			AccountID: accountID.UUID(),
			Status:    string(sharedkyc.StatusDraft),
			CreatedAt: now,
			UpdatedAt: now,
		}, documentViews, nil
	}
	return &AccountKYCProfileView{
		AccountID:        profile.AccountID,
		Status:           profile.Status,
		LegalName:        profile.LegalName,
		CountryCode:      profile.CountryCode,
		DocumentNumber:   profile.DocumentNumber,
		ResidenceAddress: profile.ResidenceAddress,
		ReviewNote:       profile.ReviewNote,
		ReviewerAccount:  profile.ReviewerAccount,
		SubmittedAt:      profile.SubmittedAt,
		ReviewedAt:       profile.ReviewedAt,
		CreatedAt:        profile.CreatedAt,
		UpdatedAt:        profile.UpdatedAt,
	}, documentViews, nil
}

func (s *Service) UpdateMyKYCProfile(ctx context.Context, cmd UpdateMyKYCProfileCmd) (*AccountKYCProfileView, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	current, err := s.repo.GetAccountKYCProfile(ctx, cmd.AccountID.UUID())
	if err != nil {
		return nil, err
	}
	nextStatus := sharedkyc.StatusDraft
	if current != nil {
		parsed, ok := sharedkyc.ParseStatus(current.Status)
		if ok {
			nextStatus = parsed
		}
	}
	var submittedAt *time.Time
	if current != nil {
		submittedAt = current.SubmittedAt
	}
	if cmd.Status != nil {
		parsed, ok := sharedkyc.ParseStatus(*cmd.Status)
		if !ok {
			return nil, apperrors.InvalidInput("KYC status is invalid")
		}
		nextStatus = parsed
		now := s.clock.Now()
		switch parsed {
		case sharedkyc.StatusSubmitted:
			submittedAt = &now
		case sharedkyc.StatusDraft:
			submittedAt = nil
		}
	}
	now := s.clock.Now()
	saved, err := s.repo.UpsertAccountKYCProfile(ctx, cmd.AccountID.UUID(), ports.AccountKYCProfilePatch{
		Status:           string(nextStatus),
		LegalName:        cmd.LegalName,
		CountryCode:      cmd.CountryCode,
		DocumentNumber:   cmd.DocumentNumber,
		ResidenceAddress: cmd.ResidenceAddress,
		ReviewNote:       nil,
		ReviewerAccount:  nil,
		SubmittedAt:      submittedAt,
		ReviewedAt:       nil,
		UpdatedAt:        now,
	})
	if err != nil {
		return nil, err
	}
	return &AccountKYCProfileView{
		AccountID:        saved.AccountID,
		Status:           saved.Status,
		LegalName:        saved.LegalName,
		CountryCode:      saved.CountryCode,
		DocumentNumber:   saved.DocumentNumber,
		ResidenceAddress: saved.ResidenceAddress,
		ReviewNote:       saved.ReviewNote,
		ReviewerAccount:  saved.ReviewerAccount,
		SubmittedAt:      saved.SubmittedAt,
		ReviewedAt:       saved.ReviewedAt,
		CreatedAt:        saved.CreatedAt,
		UpdatedAt:        saved.UpdatedAt,
	}, nil
}

func (s *Service) CreateMyKYCDocumentUpload(ctx context.Context, cmd CreateMyKYCDocumentUploadCmd) (*CreateMyKYCDocumentUploadResult, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if s.storage == nil || s.bucket == "" || s.uploads == nil {
		return nil, fault.Unavailable("KYC document upload is unavailable")
	}
	fileName := sanitizeFileName(strings.TrimSpace(cmd.FileName), "kyc-document.bin")
	documentType := strings.TrimSpace(cmd.DocumentType)
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := strings.TrimSpace(fileName)
	if cmd.Title != nil && strings.TrimSpace(*cmd.Title) != "" {
		title = strings.TrimSpace(*cmd.Title)
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
	uploadID := uuid.New()
	objectKey := strings.Join([]string{
		"accounts",
		"kyc",
		cmd.AccountID.UUID().String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		fileName,
	}, "/")
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: nil,
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    normalizeOptional(cmd.ContentType),
		SizeBytes:      sizeBytes,
		ChecksumSHA256: normalizeOptional(cmd.ChecksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create KYC object failed", fault.WithCause(err))
	}
	uploadURL, expiresAt, err := s.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign KYC upload failed", fault.WithCause(err))
	}
	if err := s.uploads.Create(ctx, &uploaddomain.Upload{
		ID:                 uploadID,
		OrganizationID:     nil,
		ObjectID:           objectID,
		CreatedByAccountID: cmd.AccountID.UUID(),
		Purpose:            uploaddomain.PurposeAccountKYCDocument,
		Status:             uploaddomain.StatusPending,
		Bucket:             object.Bucket,
		ObjectKey:          object.ObjectKey,
		FileName:           object.FileName,
		ContentType:        object.ContentType,
		DeclaredSizeBytes:  object.SizeBytes,
		ChecksumSHA256:     object.ChecksumSHA256,
		Metadata: map[string]any{
			"documentType": documentType,
			"title":        title,
		},
		ExpiresAt: &expiresAt,
		CreatedAt: now,
	}); err != nil {
		return nil, fault.Internal("Create KYC upload failed", fault.WithCause(err))
	}
	return &CreateMyKYCDocumentUploadResult{
		UploadID:     uploadID,
		ObjectID:     objectID,
		Bucket:       object.Bucket,
		ObjectKey:    object.ObjectKey,
		UploadURL:    uploadURL,
		ExpiresAt:    expiresAt,
		FileName:     fileName,
		SizeBytes:    sizeBytes,
		DocumentType: documentType,
		Title:        title,
	}, nil
}

func (s *Service) CompleteMyKYCDocumentUpload(ctx context.Context, cmd CompleteMyKYCDocumentUploadCmd) (*AccountKYCDocumentView, error) {
	if cmd.AccountID.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}
	if cmd.UploadID == uuid.Nil {
		return nil, apperrors.InvalidInput("uploadId is required")
	}
	if s.uploads == nil || s.storage == nil {
		return nil, fault.Unavailable("KYC document upload is unavailable")
	}
	upload, err := s.uploads.GetByID(ctx, cmd.UploadID)
	if err != nil {
		return nil, fault.Internal("Load KYC upload failed", fault.WithCause(err))
	}
	if upload == nil || upload.Purpose != uploaddomain.PurposeAccountKYCDocument || upload.CreatedByAccountID != cmd.AccountID.UUID() {
		return nil, fault.NotFound("KYC upload not found")
	}
	if upload.Status == uploaddomain.StatusReady {
		existing, err := s.repo.GetAccountKYCDocumentByObjectID(ctx, cmd.AccountID.UUID(), upload.ObjectID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			return nil, fault.NotFound("KYC document not found")
		}
		return accountKYCDocumentView(existing), nil
	}
	reader, err := s.storage.ReadObject(ctx, upload.Bucket, upload.ObjectKey)
	if err != nil {
		_, _ = s.uploads.MarkFailed(ctx, cmd.UploadID, "storage_object_missing", "Uploaded file is not available in storage", s.clock.Now())
		return nil, fault.Validation("Uploaded file is not available")
	}
	_ = reader.Close()
	documentType := strings.TrimSpace(metadataString(upload.Metadata, "documentType"))
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := strings.TrimSpace(metadataString(upload.Metadata, "title"))
	if title == "" {
		title = upload.FileName
	}
	now := s.clock.Now()
	record, err := s.repo.CreateAccountKYCDocument(ctx, ports.AccountKYCDocumentRecord{
		AccountID:    cmd.AccountID.UUID(),
		ObjectID:     upload.ObjectID,
		DocumentType: documentType,
		Title:        title,
		Status:       string(sharedkyc.DocumentStatusUploaded),
		CreatedAt:    now,
	})
	if err != nil {
		return nil, err
	}
	actualSize := int64Ptr(upload.DeclaredSizeBytes)
	_, err = s.uploads.MarkReady(ctx, upload.ID, actualSize, uploaddomain.ResultKindAccountKYCDocument, record.ID, now, now)
	if err != nil {
		return nil, fault.Internal("Finalize KYC upload failed", fault.WithCause(err))
	}
	return accountKYCDocumentView(record), nil
}

func accountKYCDocumentView(item *ports.AccountKYCDocumentRecord) *AccountKYCDocumentView {
	if item == nil {
		return nil
	}
	return &AccountKYCDocumentView{
		ID:              item.ID,
		AccountID:       item.AccountID,
		ObjectID:        item.ObjectID,
		DocumentType:    item.DocumentType,
		Title:           item.Title,
		Status:          item.Status,
		ReviewNote:      item.ReviewNote,
		ReviewerAccount: item.ReviewerAccount,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		ReviewedAt:      item.ReviewedAt,
	}
}

func metadataString(metadata map[string]any, key string) string {
	if len(metadata) == 0 {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
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

func buildAccountVideoObjectKey(accountID, objectID uuid.UUID, fileName string, now time.Time) string {
	return strings.Join([]string{
		"accounts",
		"videos",
		accountID.String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		sanitizeFileName(fileName, "video.mp4"),
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

func int64Ptr(value int64) *int64 {
	return &value
}

func normalizeVideoContentType(fileName, contentType string) (string, error) {
	trimmedType := strings.TrimSpace(contentType)
	if strings.HasPrefix(strings.ToLower(trimmedType), "video/") {
		return trimmedType, nil
	}

	switch strings.ToLower(filepath.Ext(strings.TrimSpace(fileName))) {
	case ".mp4":
		return "video/mp4", nil
	case ".webm":
		return "video/webm", nil
	case ".mov":
		return "video/quicktime", nil
	case ".m4v":
		return "video/x-m4v", nil
	case ".ogv":
		return "video/ogg", nil
	default:
		return "", apperrors.InvalidInput("Only video files are supported")
	}
}
