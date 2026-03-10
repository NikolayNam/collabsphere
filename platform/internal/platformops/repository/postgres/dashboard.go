package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
)

type dashboardSummaryRow struct {
	TotalAccounts          int64 `gorm:"column:total_accounts"`
	ActiveAccounts         int64 `gorm:"column:active_accounts"`
	TotalOrganizations     int64 `gorm:"column:total_organizations"`
	ActiveOrganizations    int64 `gorm:"column:active_organizations"`
	PendingUploads         int64 `gorm:"column:pending_uploads"`
	ReadyUploads           int64 `gorm:"column:ready_uploads"`
	FailedUploads          int64 `gorm:"column:failed_uploads"`
	CooperationDraft       int64 `gorm:"column:cooperation_draft"`
	CooperationSubmitted   int64 `gorm:"column:cooperation_submitted"`
	CooperationUnderReview int64 `gorm:"column:cooperation_under_review"`
	CooperationApproved    int64 `gorm:"column:cooperation_approved"`
	CooperationRejected    int64 `gorm:"column:cooperation_rejected"`
	CooperationNeedsInfo   int64 `gorm:"column:cooperation_needs_info"`
}

func (r *Repo) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	var row dashboardSummaryRow
	err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
    (SELECT count(*) FROM iam.accounts WHERE deleted_at IS NULL) AS total_accounts,
    (SELECT count(*) FROM iam.accounts WHERE deleted_at IS NULL AND is_active = true) AS active_accounts,
    (SELECT count(*) FROM org.organizations WHERE deleted_at IS NULL) AS total_organizations,
    (SELECT count(*) FROM org.organizations WHERE deleted_at IS NULL AND is_active = true) AS active_organizations,
    (SELECT count(*) FROM storage.uploads WHERE status = 'pending') AS pending_uploads,
    (SELECT count(*) FROM storage.uploads WHERE status = 'ready') AS ready_uploads,
    (SELECT count(*) FROM storage.uploads WHERE status = 'failed') AS failed_uploads,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'draft') AS cooperation_draft,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'submitted') AS cooperation_submitted,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'under_review') AS cooperation_under_review,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'approved') AS cooperation_approved,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'rejected') AS cooperation_rejected,
    (SELECT count(*) FROM org.cooperation_applications WHERE status = 'needs_info') AS cooperation_needs_info
`).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &domain.DashboardSummary{
		TotalAccounts:          row.TotalAccounts,
		ActiveAccounts:         row.ActiveAccounts,
		TotalOrganizations:     row.TotalOrganizations,
		ActiveOrganizations:    row.ActiveOrganizations,
		PendingUploads:         row.PendingUploads,
		ReadyUploads:           row.ReadyUploads,
		FailedUploads:          row.FailedUploads,
		CooperationDraft:       row.CooperationDraft,
		CooperationSubmitted:   row.CooperationSubmitted,
		CooperationUnderReview: row.CooperationUnderReview,
		CooperationApproved:    row.CooperationApproved,
		CooperationRejected:    row.CooperationRejected,
		CooperationNeedsInfo:   row.CooperationNeedsInfo,
	}, nil
}
