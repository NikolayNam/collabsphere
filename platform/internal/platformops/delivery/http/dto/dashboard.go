package dto

type DashboardSummaryInput struct{}

type DashboardSummaryResponse struct {
	Status int `json:"-"`
	Body   struct {
		TotalAccounts          int64 `json:"totalAccounts"`
		ActiveAccounts         int64 `json:"activeAccounts"`
		TotalOrganizations     int64 `json:"totalOrganizations"`
		ActiveOrganizations    int64 `json:"activeOrganizations"`
		PendingUploads         int64 `json:"pendingUploads"`
		ReadyUploads           int64 `json:"readyUploads"`
		FailedUploads          int64 `json:"failedUploads"`
		CooperationDraft       int64 `json:"cooperationDraft"`
		CooperationSubmitted   int64 `json:"cooperationSubmitted"`
		CooperationUnderReview int64 `json:"cooperationUnderReview"`
		CooperationApproved    int64 `json:"cooperationApproved"`
		CooperationRejected    int64 `json:"cooperationRejected"`
		CooperationNeedsInfo   int64 `json:"cooperationNeedsInfo"`
	}
}
