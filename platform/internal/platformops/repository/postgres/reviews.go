package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	orgdb "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationReviewQueueRow struct {
	OrganizationID           uuid.UUID  `gorm:"column:organization_id"`
	OrganizationName         string     `gorm:"column:organization_name"`
	OrganizationSlug         string     `gorm:"column:organization_slug"`
	OrganizationIsActive     bool       `gorm:"column:organization_is_active"`
	CooperationApplicationID uuid.UUID  `gorm:"column:cooperation_application_id"`
	CooperationStatus        string     `gorm:"column:cooperation_status"`
	CompanyName              *string    `gorm:"column:company_name"`
	ConfirmationEmail        *string    `gorm:"column:confirmation_email"`
	ReviewerAccountID        *uuid.UUID `gorm:"column:reviewer_account_id"`
	SubmittedAt              *time.Time `gorm:"column:submitted_at"`
	ReviewedAt               *time.Time `gorm:"column:reviewed_at"`
	CreatedAt                time.Time  `gorm:"column:created_at"`
	UpdatedAt                *time.Time `gorm:"column:updated_at"`
}

type legalDocumentReviewRow struct {
	ID                  uuid.UUID       `gorm:"column:id"`
	OrganizationID      uuid.UUID       `gorm:"column:organization_id"`
	DocumentType        string          `gorm:"column:document_type"`
	Status              string          `gorm:"column:status"`
	ObjectID            uuid.UUID       `gorm:"column:object_id"`
	Title               string          `gorm:"column:title"`
	UploadedByAccountID *uuid.UUID      `gorm:"column:uploaded_by_account_id"`
	ReviewerAccountID   *uuid.UUID      `gorm:"column:reviewer_account_id"`
	ReviewNote          *string         `gorm:"column:review_note"`
	CreatedAt           time.Time       `gorm:"column:created_at"`
	UpdatedAt           *time.Time      `gorm:"column:updated_at"`
	ReviewedAt          *time.Time      `gorm:"column:reviewed_at"`
	AnalysisID          *uuid.UUID      `gorm:"column:analysis_id"`
	AnalysisStatus      *string         `gorm:"column:analysis_status"`
	AnalysisProvider    *string         `gorm:"column:analysis_provider"`
	AnalysisSummary     *string         `gorm:"column:analysis_summary"`
	ExtractedFieldsJSON json.RawMessage `gorm:"column:extracted_fields_json"`
	DetectedType        *string         `gorm:"column:detected_document_type"`
	ConfidenceScore     *float64        `gorm:"column:confidence_score"`
	RequestedAt         *time.Time      `gorm:"column:requested_at"`
	StartedAt           *time.Time      `gorm:"column:started_at"`
	CompletedAt         *time.Time      `gorm:"column:completed_at"`
	AnalysisUpdatedAt   *time.Time      `gorm:"column:analysis_updated_at"`
	LastError           *string         `gorm:"column:last_error"`
}

func (r *Repo) ListOrganizationReviewQueue(ctx context.Context, query domain.OrganizationReviewQueueQuery) ([]domain.OrganizationReviewQueueItem, int, error) {
	db := r.dbFrom(ctx).WithContext(ctx).
		Table("org.cooperation_applications AS ca").
		Joins("JOIN org.organizations AS o ON o.id = ca.organization_id").
		Where("o.deleted_at IS NULL")

	if query.Status != nil {
		db = db.Where("ca.status = ?", *query.Status)
	}
	if query.OrganizationID != nil {
		db = db.Where("ca.organization_id = ?", *query.OrganizationID)
	}
	if query.ReviewerAccountID != nil {
		db = db.Where("ca.reviewer_account_id = ?", *query.ReviewerAccountID)
	}
	if query.Search != nil {
		search := "%" + strings.ToLower(strings.TrimSpace(*query.Search)) + "%"
		db = db.Where(
			"LOWER(o.name) LIKE ? OR LOWER(o.slug) LIKE ? OR LOWER(COALESCE(ca.company_name, '')) LIKE ? OR LOWER(COALESCE(ca.confirmation_email, '')) LIKE ?",
			search,
			search,
			search,
			search,
		)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]organizationReviewQueueRow, 0, query.Limit)
	if err := db.
		Select(`
			ca.organization_id AS organization_id,
			o.name AS organization_name,
			o.slug AS organization_slug,
			o.is_active AS organization_is_active,
			ca.id AS cooperation_application_id,
			ca.status AS cooperation_status,
			ca.company_name AS company_name,
			ca.confirmation_email AS confirmation_email,
			ca.reviewer_account_id AS reviewer_account_id,
			ca.submitted_at AS submitted_at,
			ca.reviewed_at AS reviewed_at,
			ca.created_at AS created_at,
			ca.updated_at AS updated_at`).
		Order("ca.submitted_at DESC NULLS LAST, ca.created_at DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]domain.OrganizationReviewQueueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.OrganizationReviewQueueItem{
			OrganizationID:           row.OrganizationID,
			OrganizationName:         row.OrganizationName,
			OrganizationSlug:         row.OrganizationSlug,
			OrganizationIsActive:     row.OrganizationIsActive,
			CooperationApplicationID: row.CooperationApplicationID,
			CooperationStatus:        row.CooperationStatus,
			CompanyName:              row.CompanyName,
			ConfirmationEmail:        row.ConfirmationEmail,
			ReviewerAccountID:        row.ReviewerAccountID,
			SubmittedAt:              row.SubmittedAt,
			ReviewedAt:               row.ReviewedAt,
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
		})
	}

	return items, int(total), nil
}

func (r *Repo) GetOrganizationReview(ctx context.Context, organizationID uuid.UUID) (*domain.OrganizationReviewDetail, error) {
	db := r.dbFrom(ctx).WithContext(ctx)

	var orgModel orgdb.Organization
	if err := db.Where("id = ? AND deleted_at IS NULL", organizationID).Take(&orgModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	detail := &domain.OrganizationReviewDetail{
		Organization: domain.OrganizationReviewOrganization{
			ID:           orgModel.ID,
			Name:         orgModel.Name,
			Slug:         orgModel.Slug,
			LogoObjectID: orgModel.LogoObjectID,
			Description:  orgModel.Description,
			Website:      orgModel.Website,
			PrimaryEmail: orgModel.PrimaryEmail,
			Phone:        orgModel.Phone,
			Address:      orgModel.Address,
			Industry:     orgModel.Industry,
			IsActive:     orgModel.IsActive,
			CreatedAt:    orgModel.CreatedAt,
			UpdatedAt:    orgModel.UpdatedAt,
		},
	}

	var domainModels []orgdb.OrganizationDomain
	if err := db.
		Where("organization_id = ? AND disabled_at IS NULL", organizationID).
		Order("is_primary DESC, hostname ASC").
		Find(&domainModels).Error; err != nil {
		return nil, err
	}
	detail.Domains = make([]domain.OrganizationReviewDomain, 0, len(domainModels))
	for _, item := range domainModels {
		updatedAt := item.UpdatedAt
		detail.Domains = append(detail.Domains, domain.OrganizationReviewDomain{
			ID:         item.ID,
			Hostname:   item.Hostname,
			Kind:       item.Kind,
			IsPrimary:  item.IsPrimary,
			IsVerified: item.VerifiedAt != nil,
			VerifiedAt: item.VerifiedAt,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  &updatedAt,
		})
	}

	var cooperationModel orgdb.CooperationApplication
	if err := db.Where("organization_id = ?", organizationID).Take(&cooperationModel).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		salesChannels, err := decodeSalesChannels(cooperationModel.SalesChannels)
		if err != nil {
			return nil, err
		}
		detail.CooperationApplication = &domain.OrganizationReviewCooperationApplication{
			ID:                    cooperationModel.ID,
			OrganizationID:        cooperationModel.OrganizationID,
			Status:                cooperationModel.Status,
			ConfirmationEmail:     cooperationModel.ConfirmationEmail,
			CompanyName:           cooperationModel.CompanyName,
			RepresentedCategories: cooperationModel.RepresentedCategories,
			MinimumOrderAmount:    cooperationModel.MinimumOrderAmount,
			DeliveryGeography:     cooperationModel.DeliveryGeography,
			SalesChannels:         salesChannels,
			StorefrontURL:         cooperationModel.StorefrontURL,
			ContactFirstName:      cooperationModel.ContactFirstName,
			ContactLastName:       cooperationModel.ContactLastName,
			ContactJobTitle:       cooperationModel.ContactJobTitle,
			PriceListObjectID:     cooperationModel.PriceListObjectID,
			ContactEmail:          cooperationModel.ContactEmail,
			ContactPhone:          cooperationModel.ContactPhone,
			PartnerCode:           cooperationModel.PartnerCode,
			ReviewNote:            cooperationModel.ReviewNote,
			ReviewerAccountID:     cooperationModel.ReviewerAccountID,
			SubmittedAt:           cooperationModel.SubmittedAt,
			ReviewedAt:            cooperationModel.ReviewedAt,
			CreatedAt:             cooperationModel.CreatedAt,
			UpdatedAt:             cooperationModel.UpdatedAt,
		}
	}

	rows := make([]legalDocumentReviewRow, 0)
	var err error
	rows, err = r.loadLegalDocumentReviewRows(ctx, organizationID, nil)
	if err != nil {
		return nil, err
	}

	detail.LegalDocuments = mapLegalDocumentReviewRows(rows)

	return detail, nil
}

func (r *Repo) UpdateCooperationApplicationReview(ctx context.Context, organizationID uuid.UUID, patch domain.CooperationApplicationReviewPatch) (*domain.OrganizationReviewCooperationApplication, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	updates := map[string]any{
		"status":              patch.Status,
		"review_note":         patch.ReviewNote,
		"reviewer_account_id": patch.ReviewerAccountID,
		"reviewed_at":         patch.ReviewedAt,
		"updated_at":          patch.UpdatedAt,
	}
	result := db.Table("org.cooperation_applications").Where("organization_id = ?", organizationID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	var cooperationModel orgdb.CooperationApplication
	if err := db.Where("organization_id = ?", organizationID).Take(&cooperationModel).Error; err != nil {
		return nil, err
	}
	salesChannels, err := decodeSalesChannels(cooperationModel.SalesChannels)
	if err != nil {
		return nil, err
	}
	return &domain.OrganizationReviewCooperationApplication{
		ID:                    cooperationModel.ID,
		OrganizationID:        cooperationModel.OrganizationID,
		Status:                cooperationModel.Status,
		ConfirmationEmail:     cooperationModel.ConfirmationEmail,
		CompanyName:           cooperationModel.CompanyName,
		RepresentedCategories: cooperationModel.RepresentedCategories,
		MinimumOrderAmount:    cooperationModel.MinimumOrderAmount,
		DeliveryGeography:     cooperationModel.DeliveryGeography,
		SalesChannels:         salesChannels,
		StorefrontURL:         cooperationModel.StorefrontURL,
		ContactFirstName:      cooperationModel.ContactFirstName,
		ContactLastName:       cooperationModel.ContactLastName,
		ContactJobTitle:       cooperationModel.ContactJobTitle,
		PriceListObjectID:     cooperationModel.PriceListObjectID,
		ContactEmail:          cooperationModel.ContactEmail,
		ContactPhone:          cooperationModel.ContactPhone,
		PartnerCode:           cooperationModel.PartnerCode,
		ReviewNote:            cooperationModel.ReviewNote,
		ReviewerAccountID:     cooperationModel.ReviewerAccountID,
		SubmittedAt:           cooperationModel.SubmittedAt,
		ReviewedAt:            cooperationModel.ReviewedAt,
		CreatedAt:             cooperationModel.CreatedAt,
		UpdatedAt:             cooperationModel.UpdatedAt,
	}, nil
}

func (r *Repo) UpdateLegalDocumentReview(ctx context.Context, organizationID, documentID uuid.UUID, patch domain.LegalDocumentReviewPatch) (*domain.OrganizationReviewLegalDocument, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	updates := map[string]any{
		"status":              patch.Status,
		"review_note":         patch.ReviewNote,
		"reviewer_account_id": patch.ReviewerAccountID,
		"reviewed_at":         patch.ReviewedAt,
		"updated_at":          patch.UpdatedAt,
	}
	result := db.Table("org.organization_legal_documents").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID, documentID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	rows, err := r.loadLegalDocumentReviewRows(ctx, organizationID, &documentID)
	if err != nil {
		return nil, err
	}
	items := mapLegalDocumentReviewRows(rows)
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *Repo) loadLegalDocumentReviewRows(ctx context.Context, organizationID uuid.UUID, documentID *uuid.UUID) ([]legalDocumentReviewRow, error) {
	db := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organization_legal_documents AS d").
		Select(`
			d.id AS id,
			d.organization_id AS organization_id,
			d.document_type AS document_type,
			d.status AS status,
			d.object_id AS object_id,
			d.title AS title,
			d.uploaded_by_account_id AS uploaded_by_account_id,
			d.reviewer_account_id AS reviewer_account_id,
			d.review_note AS review_note,
			d.created_at AS created_at,
			d.updated_at AS updated_at,
			d.reviewed_at AS reviewed_at,
			a.id AS analysis_id,
			a.status AS analysis_status,
			a.provider AS analysis_provider,
			a.summary AS analysis_summary,
			a.extracted_fields_json AS extracted_fields_json,
			a.detected_document_type AS detected_document_type,
			a.confidence_score AS confidence_score,
			a.requested_at AS requested_at,
			a.started_at AS started_at,
			a.completed_at AS completed_at,
			a.updated_at AS analysis_updated_at,
			a.last_error AS last_error`).
		Joins("LEFT JOIN org.organization_legal_document_analysis AS a ON a.document_id = d.id").
		Where("d.organization_id = ? AND d.deleted_at IS NULL", organizationID)
	if documentID != nil {
		db = db.Where("d.id = ?", *documentID)
	}

	rows := make([]legalDocumentReviewRow, 0, 1)
	if err := db.Order("d.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func mapLegalDocumentReviewRows(rows []legalDocumentReviewRow) []domain.OrganizationReviewLegalDocument {
	items := make([]domain.OrganizationReviewLegalDocument, 0, len(rows))
	for _, row := range rows {
		item := domain.OrganizationReviewLegalDocument{
			ID:                  row.ID,
			OrganizationID:      row.OrganizationID,
			DocumentType:        row.DocumentType,
			Status:              row.Status,
			ObjectID:            row.ObjectID,
			Title:               row.Title,
			UploadedByAccountID: row.UploadedByAccountID,
			ReviewerAccountID:   row.ReviewerAccountID,
			ReviewNote:          row.ReviewNote,
			CreatedAt:           row.CreatedAt,
			UpdatedAt:           row.UpdatedAt,
			ReviewedAt:          row.ReviewedAt,
		}
		if row.AnalysisID != nil && row.AnalysisStatus != nil && row.AnalysisProvider != nil && row.RequestedAt != nil {
			item.Analysis = &domain.OrganizationReviewLegalDocumentAnalysis{
				ID:                   *row.AnalysisID,
				DocumentID:           row.ID,
				OrganizationID:       row.OrganizationID,
				Status:               *row.AnalysisStatus,
				Provider:             *row.AnalysisProvider,
				Summary:              row.AnalysisSummary,
				ExtractedFieldsJSON:  row.ExtractedFieldsJSON,
				DetectedDocumentType: row.DetectedType,
				ConfidenceScore:      row.ConfidenceScore,
				RequestedAt:          *row.RequestedAt,
				StartedAt:            row.StartedAt,
				CompletedAt:          row.CompletedAt,
				UpdatedAt:            row.AnalysisUpdatedAt,
				LastError:            row.LastError,
			}
		}
		items = append(items, item)
	}
	return items
}

func decodeSalesChannels(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}
