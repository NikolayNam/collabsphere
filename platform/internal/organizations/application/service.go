package application

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
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

type GetCooperationApplicationQuery struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
}

type UpdateCooperationApplicationCmd struct {
	OrganizationID        domain.OrganizationID
	ActorAccountID        uuid.UUID
	ConfirmationEmail     *string
	CompanyName           *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	StorefrontURL         *string
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	PriceListObjectID     *uuid.UUID
	ClearPriceList        bool
	ContactEmail          *string
	ContactPhone          *string
	PartnerCode           *string
}

type SubmitCooperationApplicationCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
}

type CreateCooperationPriceListUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type CreateCooperationPriceListUploadResult struct {
	ObjectID  uuid.UUID
	Bucket    string
	ObjectKey string
	UploadURL string
	ExpiresAt time.Time
	FileName  string
	SizeBytes int64
}

type CreateLegalDocumentUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentType   string
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type CreateLegalDocumentUploadResult struct {
	ObjectID     uuid.UUID
	Bucket       string
	ObjectKey    string
	UploadURL    string
	ExpiresAt    time.Time
	FileName     string
	SizeBytes    int64
	DocumentType string
}

type AddOrganizationLegalDocumentCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentType   string
	ObjectID       uuid.UUID
	Title          string
}

type GetOrganizationLegalDocumentAnalysisQuery struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentID     uuid.UUID
}

type ReprocessOrganizationLegalDocumentAnalysisCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentID     uuid.UUID
}

type Service struct {
	create                  *create_organization.Handler
	getById                 *get_organization_by_id.Handler
	repo                    ports.OrganizationRepository
	memberships             memberPorts.MembershipRepository
	clock                   ports.Clock
	storage                 ports.ObjectStorage
	bucket                  string
	legalDocumentAnalyzer   ports.LegalDocumentAnalyzer
	legalDocumentAIProvider string
}

func New(repo ports.OrganizationRepository, membershipRepo memberPorts.MembershipRepository, categoryProvisioner ports.ProductCategoryProvisioner, txm sharedtx.Manager, clock ports.Clock, storage ports.ObjectStorage, bucket string, analyzer ports.LegalDocumentAnalyzer, legalDocumentAIProvider string) *Service {
	creator := create_with_owner.New(txm, repo, membershipRepo, categoryProvisioner)

	return &Service{
		create:                  create_organization.NewHandler(creator, clock),
		getById:                 get_organization_by_id.NewHandler(repo),
		repo:                    repo,
		memberships:             membershipRepo,
		clock:                   clock,
		storage:                 storage,
		bucket:                  strings.TrimSpace(bucket),
		legalDocumentAnalyzer:   analyzer,
		legalDocumentAIProvider: strings.TrimSpace(legalDocumentAIProvider),
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
	object, uploadURL, expiresAt, err := s.createOrganizationScopedUpload(ctx, cmd.OrganizationID, "organizations", []string{"logos"}, cmd.FileName, cmd.ContentType, cmd.SizeBytes, cmd.ChecksumSHA256, "logo.bin")
	if err != nil {
		return nil, err
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

func (s *Service) GetCooperationApplication(ctx context.Context, q GetCooperationApplicationQuery) (*domain.CooperationApplication, error) {
	if err := s.requireOrganizationAccess(ctx, q.OrganizationID, q.ActorAccountID, true); err != nil {
		return nil, err
	}
	application, err := s.repo.GetCooperationApplication(ctx, q.OrganizationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, apperrors.CooperationApplicationNotFound()
	}
	return application, nil
}

func (s *Service) UpdateCooperationApplication(ctx context.Context, cmd UpdateCooperationApplicationCmd) (*domain.CooperationApplication, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.ClearPriceList && cmd.PriceListObjectID != nil {
		return nil, apperrors.InvalidInput("clearPriceList and priceListObjectId cannot be used together")
	}

	application, err := s.repo.GetCooperationApplication(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		application, err = domain.NewCooperationApplication(domain.NewCooperationApplicationParams{
			ID:             uuid.New(),
			OrganizationID: cmd.OrganizationID,
			Now:            s.clock.Now(),
		})
		if err != nil {
			return nil, apperrors.InvalidInput(err.Error())
		}
	}
	if err := application.ApplyPatch(domain.CooperationApplicationPatch{
		ConfirmationEmail:     cmd.ConfirmationEmail,
		CompanyName:           cmd.CompanyName,
		RepresentedCategories: cmd.RepresentedCategories,
		MinimumOrderAmount:    cmd.MinimumOrderAmount,
		DeliveryGeography:     cmd.DeliveryGeography,
		SalesChannels:         cmd.SalesChannels,
		StorefrontURL:         cmd.StorefrontURL,
		ContactFirstName:      cmd.ContactFirstName,
		ContactLastName:       cmd.ContactLastName,
		ContactJobTitle:       cmd.ContactJobTitle,
		PriceListObjectID:     cmd.PriceListObjectID,
		ClearPriceList:        cmd.ClearPriceList,
		ContactEmail:          cmd.ContactEmail,
		ContactPhone:          cmd.ContactPhone,
		PartnerCode:           cmd.PartnerCode,
		UpdatedAt:             s.clock.Now(),
	}); err != nil {
		return nil, apperrors.InvalidInput(err.Error())
	}
	return s.repo.SaveCooperationApplication(ctx, application)
}

func (s *Service) SubmitCooperationApplication(ctx context.Context, cmd SubmitCooperationApplicationCmd) (*domain.CooperationApplication, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	application, err := s.repo.GetCooperationApplication(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, apperrors.CooperationApplicationNotFound()
	}
	documents, err := s.repo.ListOrganizationLegalDocuments(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, apperrors.InvalidInput("At least one legal document is required before submission")
	}
	if err := application.MarkSubmitted(s.clock.Now()); err != nil {
		return nil, apperrors.InvalidInput(err.Error())
	}
	return s.repo.SaveCooperationApplication(ctx, application)
}

func (s *Service) CreateCooperationPriceListUpload(ctx context.Context, cmd CreateCooperationPriceListUploadCmd) (*CreateCooperationPriceListUploadResult, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	object, uploadURL, expiresAt, err := s.createOrganizationScopedUpload(ctx, cmd.OrganizationID, "organizations", []string{"cooperation-applications", "price-lists"}, cmd.FileName, cmd.ContentType, cmd.SizeBytes, cmd.ChecksumSHA256, "price-list.xlsx")
	if err != nil {
		return nil, err
	}
	return &CreateCooperationPriceListUploadResult{
		ObjectID:  object.ID,
		Bucket:    object.Bucket,
		ObjectKey: object.ObjectKey,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
		FileName:  object.FileName,
		SizeBytes: object.SizeBytes,
	}, nil
}

func (s *Service) CreateLegalDocumentUpload(ctx context.Context, cmd CreateLegalDocumentUploadCmd) (*CreateLegalDocumentUploadResult, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	documentType := strings.TrimSpace(cmd.DocumentType)
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	object, uploadURL, expiresAt, err := s.createOrganizationScopedUpload(ctx, cmd.OrganizationID, "organizations", []string{"legal-documents", sanitizePathSegment(documentType, "other")}, cmd.FileName, cmd.ContentType, cmd.SizeBytes, cmd.ChecksumSHA256, "document.pdf")
	if err != nil {
		return nil, err
	}
	return &CreateLegalDocumentUploadResult{
		ObjectID:     object.ID,
		Bucket:       object.Bucket,
		ObjectKey:    object.ObjectKey,
		UploadURL:    uploadURL,
		ExpiresAt:    expiresAt,
		FileName:     object.FileName,
		SizeBytes:    object.SizeBytes,
		DocumentType: documentType,
	}, nil
}

func (s *Service) AddOrganizationLegalDocument(ctx context.Context, cmd AddOrganizationLegalDocumentCmd) (*domain.OrganizationLegalDocument, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	document, err := domain.NewOrganizationLegalDocument(domain.NewOrganizationLegalDocumentParams{
		ID:                  uuid.New(),
		OrganizationID:      cmd.OrganizationID,
		DocumentType:        cmd.DocumentType,
		ObjectID:            cmd.ObjectID,
		Title:               cmd.Title,
		UploadedByAccountID: &cmd.ActorAccountID,
		Now:                 s.clock.Now(),
	})
	if err != nil {
		return nil, apperrors.InvalidInput(err.Error())
	}
	created, err := s.repo.CreateOrganizationLegalDocument(ctx, document)
	if err != nil {
		return nil, err
	}
	if created != nil && s.legalDocumentAIProvider != "" {
		if err := s.repo.EnsureOrganizationLegalDocumentAnalysis(ctx, created, s.legalDocumentAIProvider, s.clock.Now()); err != nil {
			return nil, fault.Internal("Queue legal document analysis failed", fault.WithCause(err))
		}
	}
	return created, nil
}

func (s *Service) ListOrganizationLegalDocuments(ctx context.Context, organizationID domain.OrganizationID, actorAccountID uuid.UUID) ([]domain.OrganizationLegalDocument, error) {
	if err := s.requireOrganizationAccess(ctx, organizationID, actorAccountID, true); err != nil {
		return nil, err
	}
	return s.repo.ListOrganizationLegalDocuments(ctx, organizationID)
}

func (s *Service) GetOrganizationLegalDocumentAnalysis(ctx context.Context, q GetOrganizationLegalDocumentAnalysisQuery) (*domain.OrganizationLegalDocumentAnalysis, error) {
	if err := s.requireOrganizationAccess(ctx, q.OrganizationID, q.ActorAccountID, true); err != nil {
		return nil, err
	}
	analysis, err := s.repo.GetOrganizationLegalDocumentAnalysis(ctx, q.OrganizationID, q.DocumentID)
	if err != nil {
		return nil, err
	}
	if analysis == nil {
		return nil, fault.NotFound("Organization legal document analysis not found")
	}
	return analysis, nil
}

func (s *Service) ReprocessOrganizationLegalDocumentAnalysis(ctx context.Context, cmd ReprocessOrganizationLegalDocumentAnalysisCmd) (*domain.OrganizationLegalDocumentAnalysis, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if s.legalDocumentAIProvider == "" {
		return nil, fault.Unavailable("Legal document analysis is unavailable")
	}
	document, err := s.repo.GetOrganizationLegalDocumentByID(ctx, cmd.OrganizationID, cmd.DocumentID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, fault.NotFound("Organization legal document not found")
	}
	if err := s.repo.EnsureOrganizationLegalDocumentAnalysis(ctx, document, s.legalDocumentAIProvider, s.clock.Now()); err != nil {
		return nil, fault.Internal("Queue legal document analysis failed", fault.WithCause(err))
	}
	return s.repo.GetOrganizationLegalDocumentAnalysis(ctx, cmd.OrganizationID, cmd.DocumentID)
}

func (s *Service) ProcessNextLegalDocumentAnalysisJob(ctx context.Context) (bool, error) {
	if s.legalDocumentAnalyzer == nil || s.storage == nil {
		return false, nil
	}
	lease, err := s.repo.LeaseNextOrganizationLegalDocumentAnalysisJob(ctx, s.clock.Now(), 2*time.Minute)
	if err != nil {
		return false, fault.Internal("Lease legal document analysis job failed", fault.WithCause(err))
	}
	if lease == nil {
		return false, nil
	}
	body, err := s.storage.ReadObject(ctx, lease.Bucket, lease.ObjectKey)
	if err != nil {
		_ = s.repo.FailOrganizationLegalDocumentAnalysisJob(ctx, lease.JobID, lease.DocumentID, lease.Provider, err.Error(), s.clock.Now().Add(30*time.Second))
		return true, fault.Internal("Read legal document object failed", fault.WithCause(err))
	}
	defer body.Close()
	result, err := s.legalDocumentAnalyzer.Analyze(ctx, lease.FileName, lease.MimeType, body)
	if err != nil {
		_ = s.repo.FailOrganizationLegalDocumentAnalysisJob(ctx, lease.JobID, lease.DocumentID, lease.Provider, err.Error(), s.clock.Now().Add(30*time.Second))
		return true, fault.Internal("Legal document analysis failed", fault.WithCause(err))
	}
	if err := s.repo.CompleteOrganizationLegalDocumentAnalysisJob(ctx, lease.JobID, lease.DocumentID, lease.Provider, result, s.clock.Now()); err != nil {
		_ = s.repo.FailOrganizationLegalDocumentAnalysisJob(ctx, lease.JobID, lease.DocumentID, lease.Provider, err.Error(), s.clock.Now().Add(30*time.Second))
		return true, fault.Internal("Store legal document analysis failed", fault.WithCause(err))
	}
	return true, nil
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

	accountID, err := accdomain.AccountIDFromUUID(actorAccountID)
	if err != nil {
		return fault.Unauthorized("Authentication required")
	}
	membership, err := s.memberships.GetMemberByAccount(ctx, organizationID, accountID)
	if err != nil {
		return fault.Internal("Get organization membership failed", fault.WithCause(err))
	}
	if membership == nil || !membership.IsActive() || membership.IsRemoved() {
		return fault.Forbidden("Organization access denied")
	}
	if requireOwner && !membership.Role().CanManageOrganizationProfile() {
		return fault.Forbidden("Only organization owners or admins can manage organization profile")
	}
	return nil
}

func (s *Service) createOrganizationScopedUpload(ctx context.Context, organizationID domain.OrganizationID, root string, segments []string, fileName string, contentType *string, sizeBytes *int64, checksumSHA256 *string, fallbackFileName string) (ports.StorageObject, string, time.Time, error) {
	if s.storage == nil || s.bucket == "" {
		return ports.StorageObject{}, "", time.Time{}, fault.Unavailable("File upload is unavailable")
	}
	trimmedFileName := strings.TrimSpace(fileName)
	if trimmedFileName == "" {
		return ports.StorageObject{}, "", time.Time{}, apperrors.InvalidInput("fileName is required")
	}
	resolvedSizeBytes := int64(0)
	if sizeBytes != nil {
		if *sizeBytes < 0 {
			return ports.StorageObject{}, "", time.Time{}, apperrors.InvalidInput("sizeBytes must be non-negative")
		}
		resolvedSizeBytes = *sizeBytes
	}
	now := s.clock.Now()
	orgUUID := organizationID.UUID()
	objectID := uuid.New()
	pathParts := []string{root}
	pathParts = append(pathParts, segments...)
	pathParts = append(pathParts,
		orgUUID.String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		sanitizeFileName(trimmedFileName, fallbackFileName),
	)
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: &orgUUID,
		Bucket:         s.bucket,
		ObjectKey:      strings.Join(pathParts, "/"),
		FileName:       sanitizeFileName(trimmedFileName, fallbackFileName),
		ContentType:    normalizeOptional(contentType),
		SizeBytes:      resolvedSizeBytes,
		ChecksumSHA256: normalizeOptional(checksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return ports.StorageObject{}, "", time.Time{}, fault.Internal("Create organization file object failed", fault.WithCause(err))
	}
	uploadURL, expiresAt, err := s.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return ports.StorageObject{}, "", time.Time{}, fault.Internal("Presign organization file upload failed", fault.WithCause(err))
	}
	return object, uploadURL, expiresAt, nil
}

func sanitizePathSegment(value, fallback string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return fallback
	}
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return fallback
	}
	return out
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
