package domain

type DashboardSummary struct {
	TotalAccounts          int64
	ActiveAccounts         int64
	TotalOrganizations     int64
	ActiveOrganizations    int64
	PendingUploads         int64
	ReadyUploads           int64
	FailedUploads          int64
	CooperationDraft       int64
	CooperationSubmitted   int64
	CooperationUnderReview int64
	CooperationApproved    int64
	CooperationRejected    int64
	CooperationNeedsInfo   int64
}
