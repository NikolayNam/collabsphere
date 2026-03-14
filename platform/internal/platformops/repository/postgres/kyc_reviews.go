package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repo) ListKYCReviews(ctx context.Context, query domain.KYCReviewQuery) ([]domain.KYCReviewItem, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	scopeFilter := ""
	args := []any{}
	if query.Scope != nil && *query.Scope != "" {
		scopeFilter = " AND t.scope = ?"
		args = append(args, *query.Scope)
	}
	statusFilter := ""
	if query.Status != nil && *query.Status != "" {
		statusFilter = " AND t.status = ?"
		args = append(args, *query.Status)
	}

	base := fmt.Sprintf(`
WITH items AS (
  SELECT ('account:' || ap.account_id::text) AS review_id, 'account' AS scope, ap.account_id AS subject_id, ap.status, ap.kyc_level_code, l.name AS kyc_level_name, ap.legal_name, ap.country_code, ap.submitted_at, ap.reviewed_at, ap.updated_at
  FROM kyc.account_profiles ap
  LEFT JOIN kyc.levels l ON l.scope = 'account' AND l.code = ap.kyc_level_code
  UNION ALL
  SELECT ('organization:' || op.organization_id::text) AS review_id, 'organization' AS scope, op.organization_id AS subject_id, op.status, op.kyc_level_code, l.name AS kyc_level_name, op.legal_name, op.country_code, op.submitted_at, op.reviewed_at, op.updated_at
  FROM kyc.organization_profiles op
  LEFT JOIN kyc.levels l ON l.scope = 'organization' AND l.code = op.kyc_level_code
)
SELECT review_id, scope, subject_id, status, kyc_level_code, kyc_level_name, legal_name, country_code, submitted_at, reviewed_at, updated_at
FROM items t
WHERE 1=1%s%s
ORDER BY updated_at DESC, review_id DESC
LIMIT ? OFFSET ?`, scopeFilter, statusFilter)
	args = append(args, limit, offset)

	var rows []domain.KYCReviewItem
	if err := r.dbFrom(ctx).WithContext(ctx).Raw(base, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	countSQL := fmt.Sprintf(`
WITH items AS (
  SELECT 'account' AS scope, status FROM kyc.account_profiles
  UNION ALL
  SELECT 'organization' AS scope, status FROM kyc.organization_profiles
)
SELECT count(*) FROM items t WHERE 1=1%s%s`, scopeFilter, statusFilter)
	countArgs := args[:0]
	if query.Scope != nil && *query.Scope != "" {
		countArgs = append(countArgs, *query.Scope)
	}
	if query.Status != nil && *query.Status != "" {
		countArgs = append(countArgs, *query.Status)
	}
	var total int64
	if err := r.dbFrom(ctx).WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	return rows, int(total), nil
}

func (r *Repo) GetKYCReview(ctx context.Context, scope string, subjectID uuid.UUID) (*domain.KYCReviewDetail, error) {
	var row domain.KYCReviewDetail
	switch scope {
	case "account":
		err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
  ('account:' || account_id::text) AS review_id,
  'account' AS scope,
  account_id AS subject_id,
  status,
  kyc_level_code,
  (SELECT name FROM kyc.levels WHERE scope = 'account' AND code = kyc_level_code LIMIT 1) AS kyc_level_name,
  legal_name,
  country_code,
  NULL::varchar AS registration_number,
  NULL::varchar AS tax_id,
  document_number,
  residence_address,
  review_note,
  reviewer_account_id,
  submitted_at,
  reviewed_at,
  created_at,
  updated_at
FROM kyc.account_profiles
WHERE account_id = ?`, subjectID).Scan(&row).Error
		if err != nil {
			return nil, err
		}
	case "organization":
		err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
  ('organization:' || organization_id::text) AS review_id,
  'organization' AS scope,
  organization_id AS subject_id,
  status,
  kyc_level_code,
  (SELECT name FROM kyc.levels WHERE scope = 'organization' AND code = kyc_level_code LIMIT 1) AS kyc_level_name,
  legal_name,
  country_code,
  registration_number,
  tax_id,
  NULL::varchar AS document_number,
  NULL::text AS residence_address,
  review_note,
  reviewer_account_id,
  submitted_at,
  reviewed_at,
  created_at,
  updated_at
FROM kyc.organization_profiles
WHERE organization_id = ?`, subjectID).Scan(&row).Error
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	if row.SubjectID == uuid.Nil {
		return nil, nil
	}
	return &row, nil
}

func (r *Repo) ApplyKYCDecision(ctx context.Context, patch domain.KYCDecisionPatch) (*domain.KYCReviewDetail, error) {
	switch patch.Scope {
	case "account":
		if err := r.dbFrom(ctx).WithContext(ctx).Exec(`
UPDATE kyc.account_profiles
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE account_id = ?`, patch.Status, patch.ReviewNote, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID).Error; err != nil {
			return nil, err
		}
		return r.GetKYCReview(ctx, patch.Scope, patch.SubjectID)
	case "organization":
		if err := r.dbFrom(ctx).WithContext(ctx).Exec(`
UPDATE kyc.organization_profiles
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE organization_id = ?`, patch.Status, patch.ReviewNote, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID).Error; err != nil {
			return nil, err
		}
		return r.GetKYCReview(ctx, patch.Scope, patch.SubjectID)
	default:
		return nil, nil
	}
}

func (r *Repo) ListKYCDocuments(ctx context.Context, scope string, subjectID uuid.UUID) ([]domain.KYCDocumentReviewItem, error) {
	var rows []domain.KYCDocumentReviewItem
	switch scope {
	case "account":
		err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
  id,
  object_id,
  document_type,
  title,
  status,
  review_note,
  reviewer_account_id,
  created_at,
  updated_at,
  reviewed_at
FROM kyc.account_documents
WHERE account_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC`, subjectID).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	case "organization":
		err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
  id,
  object_id,
  document_type,
  title,
  status,
  review_note,
  reviewer_account_id,
  created_at,
  updated_at,
  reviewed_at
FROM kyc.organization_documents
WHERE organization_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC`, subjectID).Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	default:
		return []domain.KYCDocumentReviewItem{}, nil
	}
	return rows, nil
}

func (r *Repo) ApplyKYCDocumentDecision(ctx context.Context, patch domain.KYCDocumentDecisionPatch) (*domain.KYCDocumentReviewItem, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	var row domain.KYCDocumentReviewItem
	switch patch.Scope {
	case "account":
		res := db.Exec(`
UPDATE kyc.account_documents
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE account_id = ? AND id = ? AND deleted_at IS NULL`,
			patch.Status, patch.ReviewNote, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID, patch.DocumentID)
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return nil, nil
		}
		if err := r.recomputeAccountKYCProfileStatus(ctx, patch); err != nil {
			return nil, err
		}
		err := db.Raw(`
SELECT
  id,
  object_id,
  document_type,
  title,
  status,
  review_note,
  reviewer_account_id,
  created_at,
  updated_at,
  reviewed_at
FROM kyc.account_documents
WHERE account_id = ? AND id = ? AND deleted_at IS NULL`,
			patch.SubjectID, patch.DocumentID).Scan(&row).Error
		if err != nil {
			return nil, err
		}
	case "organization":
		res := db.Exec(`
UPDATE kyc.organization_documents
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE organization_id = ? AND id = ? AND deleted_at IS NULL`,
			patch.Status, patch.ReviewNote, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID, patch.DocumentID)
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return nil, nil
		}
		if err := r.recomputeOrganizationKYCProfileStatus(ctx, patch); err != nil {
			return nil, err
		}
		err := db.Raw(`
SELECT
  id,
  object_id,
  document_type,
  title,
  status,
  review_note,
  reviewer_account_id,
  created_at,
  updated_at,
  reviewed_at
FROM kyc.organization_documents
WHERE organization_id = ? AND id = ? AND deleted_at IS NULL`,
			patch.SubjectID, patch.DocumentID).Scan(&row).Error
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	if row.ID == uuid.Nil {
		return nil, nil
	}
	return &row, nil
}

func (r *Repo) AppendKYCReviewEvent(ctx context.Context, event domain.KYCReviewEvent) error {
	return r.dbFrom(ctx).WithContext(ctx).Exec(`
INSERT INTO kyc.review_events (id, scope, subject_id, decision, reason, reviewer_account_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.Scope, event.SubjectID, event.Decision, event.Reason, event.ReviewerAccountID, event.CreatedAt).Error
}

func (r *Repo) ListKYCReviewEvents(ctx context.Context, scope string, subjectID uuid.UUID, limit int) ([]domain.KYCReviewEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var rows []domain.KYCReviewEvent
	err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT id, scope, subject_id, decision, reason, reviewer_account_id, created_at
FROM kyc.review_events
WHERE scope = ? AND subject_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?`, scope, subjectID, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repo) recomputeAccountKYCProfileStatus(ctx context.Context, patch domain.KYCDocumentDecisionPatch) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	type counters struct {
		Total    int64
		Verified int64
		Rejected int64
	}
	var c counters
	if err := db.Raw(`
SELECT
  count(*) AS total,
  count(*) FILTER (WHERE status = 'verified') AS verified,
  count(*) FILTER (WHERE status = 'rejected') AS rejected
FROM kyc.account_documents
WHERE account_id = ? AND deleted_at IS NULL`, patch.SubjectID).Scan(&c).Error; err != nil {
		return err
	}
	status := ""
	note := patch.ReviewNote
	switch {
	case c.Rejected > 0:
		status = "needs_info"
	case c.Total > 0 && c.Verified == c.Total:
		status = "approved"
		note = nil
	case c.Total > 0:
		status = "in_review"
		note = nil
	default:
		return nil
	}
	return db.Exec(`
UPDATE kyc.account_profiles
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE account_id = ?`,
		status, note, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID).Error
}

func (r *Repo) recomputeOrganizationKYCProfileStatus(ctx context.Context, patch domain.KYCDocumentDecisionPatch) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	type counters struct {
		Total    int64
		Verified int64
		Rejected int64
	}
	var c counters
	if err := db.Raw(`
SELECT
  count(*) AS total,
  count(*) FILTER (WHERE status = 'verified') AS verified,
  count(*) FILTER (WHERE status = 'rejected') AS rejected
FROM kyc.organization_documents
WHERE organization_id = ? AND deleted_at IS NULL`, patch.SubjectID).Scan(&c).Error; err != nil {
		return err
	}
	status := ""
	note := patch.ReviewNote
	switch {
	case c.Rejected > 0:
		status = "needs_info"
	case c.Total > 0 && c.Verified == c.Total:
		status = "approved"
		note = nil
	case c.Total > 0:
		status = "in_review"
		note = nil
	default:
		return nil
	}
	return db.Exec(`
UPDATE kyc.organization_profiles
SET status = ?, review_note = ?, reviewer_account_id = ?, reviewed_at = ?, updated_at = ?
WHERE organization_id = ?`,
		status, note, patch.ReviewerAccountID, patch.ReviewedAt, patch.UpdatedAt, patch.SubjectID).Error
}

func (r *Repo) ListKYCLevels(ctx context.Context, scope *string) ([]domain.KYCLevel, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	type levelRow struct {
		ID        uuid.UUID `gorm:"column:id"`
		Scope     string    `gorm:"column:scope"`
		Code      string    `gorm:"column:code"`
		Name      string    `gorm:"column:name"`
		Rank      int       `gorm:"column:rank"`
		IsActive  bool      `gorm:"column:is_active"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}
	query := db.Table("kyc.levels").
		Select("id, scope, code, name, rank, is_active, created_at, updated_at")
	if scope != nil && *scope != "" {
		query = query.Where("scope = ?", *scope)
	}
	var rows []levelRow
	if err := query.Order("scope ASC, rank ASC, code ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	levelIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		levelIDs = append(levelIDs, row.ID)
	}
	requirementsByLevel, err := r.listLevelRequirements(ctx, levelIDs)
	if err != nil {
		return nil, err
	}
	items := make([]domain.KYCLevel, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.KYCLevel{
			ID:                    row.ID,
			Scope:                 row.Scope,
			Code:                  row.Code,
			Name:                  row.Name,
			Rank:                  row.Rank,
			IsActive:              row.IsActive,
			RequiredDocumentTypes: requirementsByLevel[row.ID],
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) UpsertKYCLevel(ctx context.Context, level domain.KYCLevel) (*domain.KYCLevel, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	var result *domain.KYCLevel
	err := db.Transaction(func(tx *gorm.DB) error {
		if level.ID == uuid.Nil {
			level.ID = uuid.New()
		}
		if err := tx.Exec(`
INSERT INTO kyc.levels (id, scope, code, name, rank, is_active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
  scope = EXCLUDED.scope,
  code = EXCLUDED.code,
  name = EXCLUDED.name,
  rank = EXCLUDED.rank,
  is_active = EXCLUDED.is_active,
  updated_at = EXCLUDED.updated_at`,
			level.ID, level.Scope, level.Code, level.Name, level.Rank, level.IsActive, level.CreatedAt, level.UpdatedAt).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM kyc.level_requirements WHERE level_id = ?`, level.ID).Error; err != nil {
			return err
		}
		for _, requirement := range level.RequiredDocumentTypes {
			if err := tx.Exec(`
INSERT INTO kyc.level_requirements (level_id, document_type, min_count, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)`,
				level.ID, requirement.DocumentType, requirement.MinCount, level.UpdatedAt, level.UpdatedAt).Error; err != nil {
				return err
			}
		}
		item, err := r.getKYCLevelByID(tx, level.ID)
		if err != nil {
			return err
		}
		result = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repo) DeleteKYCLevel(ctx context.Context, levelID uuid.UUID) error {
	return r.dbFrom(ctx).WithContext(ctx).Exec(`DELETE FROM kyc.levels WHERE id = ?`, levelID).Error
}

func (r *Repo) EvaluateAndAssignKYCLevel(ctx context.Context, scope string, subjectID, actorAccountID uuid.UUID, now time.Time) (*domain.KYCLevelAssignment, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	level, err := r.matchHighestLevel(ctx, scope, subjectID)
	if err != nil {
		return nil, err
	}
	var code *string
	var name *string
	var assignedAt *time.Time
	var id *uuid.UUID
	if level != nil {
		code = &level.Code
		name = &level.Name
		assignedAt = &now
		id = &level.ID
	}
	switch scope {
	case "account":
		if err := db.Exec(`
UPDATE kyc.account_profiles
SET kyc_level_code = ?, kyc_level_assigned_at = ?, reviewer_account_id = ?, updated_at = ?
WHERE account_id = ?`, code, assignedAt, actorAccountID, now, subjectID).Error; err != nil {
			return nil, err
		}
	case "organization":
		if err := db.Exec(`
UPDATE kyc.organization_profiles
SET kyc_level_code = ?, kyc_level_assigned_at = ?, reviewer_account_id = ?, updated_at = ?
WHERE organization_id = ?`, code, assignedAt, actorAccountID, now, subjectID).Error; err != nil {
			return nil, err
		}
	default:
		return &domain.KYCLevelAssignment{}, nil
	}
	return &domain.KYCLevelAssignment{
		LevelID:   id,
		LevelCode: code,
		LevelName: name,
		IssuedAt:  assignedAt,
	}, nil
}

func (r *Repo) listLevelRequirements(ctx context.Context, levelIDs []uuid.UUID) (map[uuid.UUID][]domain.KYCLevelRequirement, error) {
	result := make(map[uuid.UUID][]domain.KYCLevelRequirement, len(levelIDs))
	if len(levelIDs) == 0 {
		return result, nil
	}
	type requirementRow struct {
		LevelID      uuid.UUID `gorm:"column:level_id"`
		DocumentType string    `gorm:"column:document_type"`
		MinCount     int       `gorm:"column:min_count"`
	}
	var rows []requirementRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.level_requirements").
		Select("level_id, document_type, min_count").
		Where("level_id IN ?", levelIDs).
		Order("document_type ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.LevelID] = append(result[row.LevelID], domain.KYCLevelRequirement{
			DocumentType: row.DocumentType,
			MinCount:     row.MinCount,
		})
	}
	return result, nil
}

func (r *Repo) getKYCLevelByID(tx *gorm.DB, levelID uuid.UUID) (*domain.KYCLevel, error) {
	type levelRow struct {
		ID        uuid.UUID `gorm:"column:id"`
		Scope     string    `gorm:"column:scope"`
		Code      string    `gorm:"column:code"`
		Name      string    `gorm:"column:name"`
		Rank      int       `gorm:"column:rank"`
		IsActive  bool      `gorm:"column:is_active"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}
	var row levelRow
	if err := tx.Raw(`SELECT id, scope, code, name, rank, is_active, created_at, updated_at FROM kyc.levels WHERE id = ?`, levelID).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, nil
	}
	type requirementRow struct {
		DocumentType string `gorm:"column:document_type"`
		MinCount     int    `gorm:"column:min_count"`
	}
	var reqRows []requirementRow
	if err := tx.Raw(`SELECT document_type, min_count FROM kyc.level_requirements WHERE level_id = ? ORDER BY document_type ASC`, row.ID).Scan(&reqRows).Error; err != nil {
		return nil, err
	}
	requirements := make([]domain.KYCLevelRequirement, 0, len(reqRows))
	for _, req := range reqRows {
		requirements = append(requirements, domain.KYCLevelRequirement{
			DocumentType: req.DocumentType,
			MinCount:     req.MinCount,
		})
	}
	return &domain.KYCLevel{
		ID:                    row.ID,
		Scope:                 row.Scope,
		Code:                  row.Code,
		Name:                  row.Name,
		Rank:                  row.Rank,
		IsActive:              row.IsActive,
		RequiredDocumentTypes: requirements,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}, nil
}

func (r *Repo) matchHighestLevel(ctx context.Context, scope string, subjectID uuid.UUID) (*domain.KYCLevel, error) {
	db := r.dbFrom(ctx).WithContext(ctx)
	docTable := ""
	subjectColumn := ""
	switch scope {
	case "account":
		docTable = "kyc.account_documents"
		subjectColumn = "account_id"
	case "organization":
		docTable = "kyc.organization_documents"
		subjectColumn = "organization_id"
	default:
		return nil, nil
	}
	sql := fmt.Sprintf(`
SELECT l.id, l.scope, l.code, l.name, l.rank, l.is_active, l.created_at, l.updated_at
FROM kyc.levels l
WHERE l.scope = ? AND l.is_active = true
  AND NOT EXISTS (
      SELECT 1
      FROM kyc.level_requirements r
      LEFT JOIN (
          SELECT document_type, count(*) AS cnt
          FROM %s
          WHERE %s = ? AND status = 'verified' AND deleted_at IS NULL
          GROUP BY document_type
      ) d ON d.document_type = r.document_type
      WHERE r.level_id = l.id AND COALESCE(d.cnt, 0) < r.min_count
  )
ORDER BY l.rank DESC, l.updated_at DESC, l.code ASC
LIMIT 1`, docTable, subjectColumn)
	type row struct {
		ID        uuid.UUID `gorm:"column:id"`
		Scope     string    `gorm:"column:scope"`
		Code      string    `gorm:"column:code"`
		Name      string    `gorm:"column:name"`
		Rank      int       `gorm:"column:rank"`
		IsActive  bool      `gorm:"column:is_active"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}
	var matched row
	if err := db.Raw(sql, scope, subjectID).Scan(&matched).Error; err != nil {
		return nil, err
	}
	if matched.ID == uuid.Nil {
		return nil, nil
	}
	reqs, err := r.listLevelRequirements(ctx, []uuid.UUID{matched.ID})
	if err != nil {
		return nil, err
	}
	return &domain.KYCLevel{
		ID:                    matched.ID,
		Scope:                 matched.Scope,
		Code:                  matched.Code,
		Name:                  matched.Name,
		Rank:                  matched.Rank,
		IsActive:              matched.IsActive,
		RequiredDocumentTypes: reqs[matched.ID],
		CreatedAt:             matched.CreatedAt,
		UpdatedAt:             matched.UpdatedAt,
	}, nil
}
