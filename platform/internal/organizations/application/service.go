package application

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	accesspolicy "github.com/NikolayNam/collabsphere/internal/iam/access"
	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization"
	create_with_owner "github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization_with_owner"
	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/get_organization_by_id"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploadports "github.com/NikolayNam/collabsphere/internal/uploads/application/ports"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
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
	Domains        *[]domain.OrganizationDomainDraft
}

type CreateOrganizationLogoUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type UploadOrganizationLogoCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    string
	SizeBytes      int64
	Body           io.Reader
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
	PriceListStatus       *string
	ClearPriceList        bool
	ContactEmail          *string
	ContactPhone          *string
	PartnerCode           *string
}

type SubmitCooperationApplicationCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
}

type PublishAllCatalogCmd struct {
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

type UploadCooperationPriceListCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    string
	SizeBytes      int64
	Body           io.Reader
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
	Title          string
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type UploadOrganizationLegalDocumentCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentType   string
	Title          string
	FileName       string
	ContentType    string
	SizeBytes      int64
	Body           io.Reader
}

type CreateLegalDocumentUploadResult struct {
	UploadID     uuid.UUID
	ObjectID     uuid.UUID
	Bucket       string
	ObjectKey    string
	UploadURL    string
	ExpiresAt    time.Time
	FileName     string
	SizeBytes    int64
	DocumentType string
}

type CompleteLegalDocumentUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	UploadID       uuid.UUID
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

type GetOrganizationLegalDocumentVerificationQuery struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentID     uuid.UUID
}

type GetOrganizationKYCRequirementsQuery struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
}

type MyOrganizationView struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	LogoObjectID   *uuid.UUID
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	MembershipRole string
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
	roleResolver            memberPorts.RoleResolver
	catalogPublisher        ports.CatalogPublisher
	tx                      sharedtx.Manager
	clock                   ports.Clock
	storage                 ports.ObjectStorage
	bucket                  string
	uploads                 uploadports.Repository
	legalDocumentAnalyzer   ports.LegalDocumentAnalyzer
	legalDocumentAIProvider string
}

const publicDirectoryKYCLevelCode = "public_directory_org_verified"

func New(repo ports.OrganizationRepository, membershipRepo memberPorts.MembershipRepository, roleResolver memberPorts.RoleResolver, categoryProvisioner ports.ProductCategoryProvisioner, catalogPublisher ports.CatalogPublisher, txm sharedtx.Manager, clock ports.Clock, storage ports.ObjectStorage, bucket string, analyzer ports.LegalDocumentAnalyzer, legalDocumentAIProvider string, uploads uploadports.Repository) *Service {
	creator := create_with_owner.New(txm, repo, membershipRepo, categoryProvisioner)

	return &Service{
		create:                  create_organization.NewHandler(creator, clock),
		getById:                 get_organization_by_id.NewHandler(repo),
		repo:                    repo,
		memberships:             membershipRepo,
		roleResolver:            roleResolver,
		catalogPublisher:        catalogPublisher,
		tx:                      txm,
		clock:                   clock,
		storage:                 storage,
		bucket:                  strings.TrimSpace(bucket),
		uploads:                 uploads,
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

func (s *Service) GetOrganizationByHost(ctx context.Context, rawHost string) (*domain.Organization, error) {
	host, err := domain.NormalizeOrganizationHostname(rawHost)
	if err != nil {
		return nil, apperrors.InvalidInput(err.Error())
	}
	return s.repo.GetByHostname(ctx, host)
}

func (s *Service) ListMyOrganizations(ctx context.Context, actorAccountID uuid.UUID) ([]MyOrganizationView, error) {
	if actorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required")
	}
	items, err := s.repo.ListByAccount(ctx, actorAccountID)
	if err != nil {
		return nil, err
	}
	out := make([]MyOrganizationView, 0, len(items))
	for _, item := range items {
		out = append(out, MyOrganizationView{
			ID:             item.ID,
			Name:           item.Name,
			Slug:           item.Slug,
			LogoObjectID:   item.LogoObjectID,
			IsActive:       item.IsActive,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			MembershipRole: item.MembershipRole,
		})
	}
	return out, nil
}

func (s *Service) ListPublicKYCDirectoryOrganizations(ctx context.Context, limit int) ([]ports.PublicKYCDirectoryOrganization, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	return s.repo.ListPublicKYCDirectoryOrganizations(ctx, publicDirectoryKYCLevelCode, limit)
}

func (s *Service) ListOrganizationDomains(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationDomain, error) {
	return s.repo.ListDomains(ctx, organizationID)
}

func (s *Service) UpdateOrganizationProfile(ctx context.Context, cmd UpdateOrganizationProfileCmd) (*domain.Organization, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.ClearLogo && cmd.LogoObjectID != nil {
		return nil, apperrors.InvalidInput("clearLogo and logoObjectId cannot be used together")
	}

	now := s.clock.Now()
	var updated *domain.Organization
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := s.repo.UpdateProfile(ctx, cmd.OrganizationID, domain.OrganizationProfilePatch{
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
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}
		if current == nil {
			return nil
		}
		if cmd.Domains != nil {
			existingDomains, err := s.repo.ListDomains(ctx, cmd.OrganizationID)
			if err != nil {
				return err
			}
			domainsToStore, err := domain.BuildOrganizationDomains(cmd.OrganizationID, *cmd.Domains, existingDomains, now)
			if err != nil {
				return apperrors.InvalidInput(err.Error())
			}
			if _, err := s.repo.ReplaceDomains(ctx, cmd.OrganizationID, domainsToStore, now); err != nil {
				return err
			}
		}
		updated = current
		return nil
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

func (s *Service) UploadOrganizationLogo(ctx context.Context, cmd UploadOrganizationLogoCmd) (*domain.Organization, error) {
	if cmd.Body == nil {
		return nil, apperrors.InvalidInput("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, apperrors.InvalidInput("file size must be non-negative")
	}
	contentType := cmd.ContentType
	sizeBytes := cmd.SizeBytes
	upload, err := s.CreateOrganizationLogoUpload(ctx, CreateOrganizationLogoUploadCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		FileName:       cmd.FileName,
		ContentType:    &contentType,
		SizeBytes:      &sizeBytes,
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	if err := s.storage.PutObject(ctx, upload.Bucket, upload.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload organization logo failed", fault.WithCause(err))
	}
	return s.UpdateOrganizationProfile(ctx, UpdateOrganizationProfileCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		LogoObjectID:   &upload.ObjectID,
	})
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
		PriceListStatus:       cmd.PriceListStatus,
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

func (s *Service) PublishAllCatalog(ctx context.Context, cmd PublishAllCatalogCmd) error {
	if err := s.requireOrganizationCatalogAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return err
	}

	categories, err := s.catalogPublisher.ListProductCategories(ctx, cmd.OrganizationID)
	if err != nil {
		return fault.Internal("List product categories failed", fault.WithCause(err))
	}
	products, err := s.catalogPublisher.ListProducts(ctx, cmd.OrganizationID)
	if err != nil {
		return fault.Internal("List products failed", fault.WithCause(err))
	}
	cooperation, err := s.repo.GetCooperationApplication(ctx, cmd.OrganizationID)
	if err != nil {
		return err
	}

	hasPriceList := cooperation != nil && cooperation.PriceListObjectID() != nil
	if len(categories) == 0 && len(products) == 0 && !hasPriceList {
		return apperrors.InvalidInput("Nothing to publish: add categories, products, or a price list first")
	}

	now := s.clock.Now()
	published := string(catalogdomain.ProductCategoryStatusPublished)
	productPublished := string(catalogdomain.ProductStatusPublished)
	priceListPublished := string(domain.CooperationPriceListStatusPublished)

	for i := range categories {
		c := &categories[i]
		st := string(c.Status())
		if st != "verified" && st != "published" {
			return apperrors.InvalidInput("Not all categories are verified; cannot auto-publish")
		}
		updatedAt := now
		if c.UpdatedAt() != nil {
			updatedAt = *c.UpdatedAt()
		}
		updated, err := catalogdomain.RehydrateProductCategory(catalogdomain.RehydrateProductCategoryParams{
			ID:             c.ID(),
			OrganizationID: c.OrganizationID(),
			ParentID:       c.ParentID(),
			TemplateID:     c.TemplateID(),
			Status:         published,
			Code:           c.Code(),
			Name:           c.Name(),
			SortOrder:      c.SortOrder(),
			CreatedAt:      c.CreatedAt(),
			UpdatedAt:      updatedAt,
		})
		if err != nil {
			return apperrors.InvalidInput(err.Error())
		}
		if err := s.catalogPublisher.UpdateProductCategory(ctx, updated); err != nil {
			return fault.Internal("Update product category failed", fault.WithCause(err))
		}
	}

	for i := range products {
		p := &products[i]
		st := string(p.Status())
		if st != "verified" && st != "published" {
			return apperrors.InvalidInput("Not all products are verified; cannot auto-publish")
		}
		updated, err := catalogdomain.RehydrateProduct(catalogdomain.RehydrateProductParams{
			ID:             p.ID(),
			OrganizationID: p.OrganizationID(),
			CategoryID:     p.CategoryID(),
			Status:         productPublished,
			Name:           p.Name(),
			Description:    p.Description(),
			SKU:           p.SKU(),
			PriceAmount:   p.PriceAmount(),
			CurrencyCode:  p.CurrencyCode(),
			IsActive:      p.IsActive(),
			CreatedAt:     p.CreatedAt(),
			UpdatedAt:     now,
		})
		if err != nil {
			return apperrors.InvalidInput(err.Error())
		}
		if err := s.catalogPublisher.UpdateProduct(ctx, updated); err != nil {
			return fault.Internal("Update product failed", fault.WithCause(err))
		}
	}

	if hasPriceList {
		plSt := string(cooperation.PriceListStatus())
		if plSt != "verified" && plSt != "published" {
			return apperrors.InvalidInput("Price list is not verified; cannot auto-publish")
		}
		if err := cooperation.ApplyPatch(domain.CooperationApplicationPatch{
			PriceListStatus: &priceListPublished,
			UpdatedAt:       now,
		}); err != nil {
			return apperrors.InvalidInput(err.Error())
		}
		if _, err := s.repo.SaveCooperationApplication(ctx, cooperation); err != nil {
			return err
		}
	}

	return nil
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

func (s *Service) UploadCooperationPriceList(ctx context.Context, cmd UploadCooperationPriceListCmd) (*domain.CooperationApplication, error) {
	if cmd.Body == nil {
		return nil, apperrors.InvalidInput("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, apperrors.InvalidInput("file size must be non-negative")
	}
	contentType := cmd.ContentType
	sizeBytes := cmd.SizeBytes
	upload, err := s.CreateCooperationPriceListUpload(ctx, CreateCooperationPriceListUploadCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		FileName:       cmd.FileName,
		ContentType:    &contentType,
		SizeBytes:      &sizeBytes,
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	if err := s.storage.PutObject(ctx, upload.Bucket, upload.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload cooperation price list failed", fault.WithCause(err))
	}
	priceListStatus := string(domain.CooperationPriceListStatusValidating)
	return s.UpdateCooperationApplication(ctx, UpdateCooperationApplicationCmd{
		OrganizationID:    cmd.OrganizationID,
		ActorAccountID:    cmd.ActorAccountID,
		PriceListObjectID: &upload.ObjectID,
		PriceListStatus:   &priceListStatus,
	})
}
func (s *Service) CreateLegalDocumentUpload(ctx context.Context, cmd CreateLegalDocumentUploadCmd) (*CreateLegalDocumentUploadResult, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if s.uploads == nil {
		return nil, fault.Unavailable("Upload tracking is unavailable")
	}
	documentType := strings.TrimSpace(cmd.DocumentType)
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := normalizeLegalDocumentTitle(cmd.Title, cmd.FileName)
	createdAt := s.clock.Now()
	var result *CreateLegalDocumentUploadResult
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		object, uploadURL, expiresAt, err := s.createOrganizationScopedUpload(ctx, cmd.OrganizationID, "organizations", []string{"legal-documents", sanitizePathSegment(documentType, "other")}, cmd.FileName, cmd.ContentType, cmd.SizeBytes, cmd.ChecksumSHA256, "document.pdf")
		if err != nil {
			return err
		}
		organizationID := cmd.OrganizationID.UUID()
		upload := &uploaddomain.Upload{
			ID:                 uuid.New(),
			OrganizationID:     &organizationID,
			ObjectID:           object.ID,
			CreatedByAccountID: cmd.ActorAccountID,
			Purpose:            uploaddomain.PurposeOrganizationLegalDocument,
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
			CreatedAt: createdAt,
		}
		if err := s.uploads.Create(ctx, upload); err != nil {
			return err
		}
		result = &CreateLegalDocumentUploadResult{
			UploadID:     upload.ID,
			ObjectID:     object.ID,
			Bucket:       object.Bucket,
			ObjectKey:    object.ObjectKey,
			UploadURL:    uploadURL,
			ExpiresAt:    expiresAt,
			FileName:     object.FileName,
			SizeBytes:    object.SizeBytes,
			DocumentType: documentType,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) UploadOrganizationLegalDocument(ctx context.Context, cmd UploadOrganizationLegalDocumentCmd) (*domain.OrganizationLegalDocument, error) {
	if cmd.Body == nil {
		return nil, apperrors.InvalidInput("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, apperrors.InvalidInput("file size must be non-negative")
	}
	contentType := cmd.ContentType
	sizeBytes := cmd.SizeBytes
	upload, err := s.CreateLegalDocumentUpload(ctx, CreateLegalDocumentUploadCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		DocumentType:   cmd.DocumentType,
		Title:          cmd.Title,
		FileName:       cmd.FileName,
		ContentType:    &contentType,
		SizeBytes:      &sizeBytes,
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	if err := s.storage.PutObject(ctx, upload.Bucket, upload.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload organization legal document failed", fault.WithCause(err))
	}
	return s.CompleteLegalDocumentUpload(ctx, CompleteLegalDocumentUploadCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		UploadID:       upload.UploadID,
	})
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
	resolved, err := s.roleResolver.ResolveRoleForPermissions(ctx, organizationID, string(membership.Role()))
	if err != nil || resolved == "" {
		return fault.Forbidden("Organization access denied")
	}
	if requireOwner && !accesspolicy.HasOrganizationPermission(resolved, accesspolicy.PermissionOrganizationManageProfile) {
		return fault.Forbidden("Only organization owners or admins can manage organization profile")
	}
	return nil
}

func (s *Service) requireOrganizationCatalogAccess(ctx context.Context, organizationID domain.OrganizationID, actorAccountID uuid.UUID) error {
	if organizationID.IsZero() {
		return apperrors.InvalidInput("Organization is required")
	}
	if actorAccountID == uuid.Nil {
		return fault.Unauthorized("Authentication required")
	}
	org, err := s.repo.GetByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if org == nil {
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
	resolved, err := s.roleResolver.ResolveRoleForPermissions(ctx, organizationID, string(membership.Role()))
	if err != nil || resolved == "" {
		return fault.Forbidden("Organization access denied")
	}
	if !accesspolicy.HasOrganizationPermission(resolved, accesspolicy.PermissionOrganizationManageCatalog) {
		return fault.Forbidden("Only organization owners, admins, or catalog managers can manage catalog")
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

type OrganizationVideoView struct {
	ID          uuid.UUID
	ObjectID    uuid.UUID
	FileName    string
	ContentType *string
	SizeBytes   int64
	CreatedAt   time.Time
	UploadedBy  *uuid.UUID
	SortOrder   int64
}

type UploadOrganizationVideoCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	FileName       string
	ContentType    string
	SizeBytes      int64
	Body           io.Reader
}

func (s *Service) UploadOrganizationVideo(ctx context.Context, cmd UploadOrganizationVideoCmd) (*ports.OrganizationVideoRecord, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.Body == nil {
		return nil, apperrors.InvalidInput("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, apperrors.InvalidInput("file size must be non-negative")
	}
	contentType, err := normalizeOrganizationVideoContentType(cmd.FileName, cmd.ContentType)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	orgUUID := cmd.OrganizationID.UUID()
	objectID := uuid.New()
	fileName := sanitizeFileName(cmd.FileName, "organization-video.mp4")
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: &orgUUID,
		Bucket:         s.bucket,
		ObjectKey:      strings.Join([]string{"organizations", "videos", orgUUID.String(), now.UTC().Format("2006"), now.UTC().Format("01"), now.UTC().Format("02"), objectID.String(), fileName}, "/"),
		FileName:       fileName,
		ContentType:    &contentType,
		SizeBytes:      cmd.SizeBytes,
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create organization video object failed", fault.WithCause(err))
	}
	if err := s.storage.PutObject(ctx, object.Bucket, object.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, fault.Internal("Upload organization video failed", fault.WithCause(err))
	}
	return s.repo.CreateOrganizationVideo(ctx, orgUUID, objectID, &cmd.ActorAccountID, now)
}

func (s *Service) ListOrganizationVideos(ctx context.Context, organizationID domain.OrganizationID, actorAccountID uuid.UUID) ([]OrganizationVideoView, error) {
	if err := s.requireOrganizationAccess(ctx, organizationID, actorAccountID, false); err != nil {
		return nil, err
	}
	items, err := s.repo.ListOrganizationVideos(ctx, organizationID.UUID())
	if err != nil {
		return nil, err
	}
	out := make([]OrganizationVideoView, 0, len(items))
	for _, item := range items {
		out = append(out, OrganizationVideoView{
			ID:          item.ID,
			ObjectID:    item.ObjectID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			CreatedAt:   item.CreatedAt,
			UploadedBy:  item.UploadedBy,
			SortOrder:   item.SortOrder,
		})
	}
	return out, nil
}

func (s *Service) ListOrganizationVideoObjectIDs(ctx context.Context, organizationID domain.OrganizationID) ([]uuid.UUID, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization is required")
	}
	return s.repo.ListOrganizationVideoObjectIDs(ctx, organizationID.UUID())
}

func normalizeOrganizationVideoContentType(fileName, contentType string) (string, error) {
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
