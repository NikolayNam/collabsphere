package kyc

import "strings"

type Scope string

type Status string

type DocumentStatus string

type Decision string

const (
	ScopeAccount      Scope = "account"
	ScopeOrganization Scope = "organization"
)

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusInReview  Status = "in_review"
	StatusNeedsInfo Status = "needs_info"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
)

const (
	DocumentStatusPendingUpload DocumentStatus = "pending_upload"
	DocumentStatusUploaded      DocumentStatus = "uploaded"
	DocumentStatusVerified      DocumentStatus = "verified"
	DocumentStatusRejected      DocumentStatus = "rejected"
)

const (
	DecisionApprove     Decision = "approve"
	DecisionReject      Decision = "reject"
	DecisionRequestInfo Decision = "request_info"
)

func ParseStatus(raw string) (Status, bool) {
	value := Status(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case StatusDraft, StatusSubmitted, StatusInReview, StatusNeedsInfo, StatusApproved, StatusRejected:
		return value, true
	default:
		return "", false
	}
}

func ParseDecision(raw string) (Decision, bool) {
	value := Decision(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case DecisionApprove, DecisionReject, DecisionRequestInfo:
		return value, true
	default:
		return "", false
	}
}

func ParseScope(raw string) (Scope, bool) {
	value := Scope(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case ScopeAccount, ScopeOrganization:
		return value, true
	default:
		return "", false
	}
}
