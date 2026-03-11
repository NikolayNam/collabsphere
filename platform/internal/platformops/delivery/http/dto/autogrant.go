package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListAutoGrantRulesInput struct{}

type CreateAutoGrantRuleInput struct {
	Body struct {
		Role       string `json:"role" required:"true" doc:"Platform role to auto-grant. Supported values: platform_admin, support_operator, review_operator."`
		MatchType  string `json:"matchType" required:"true" doc:"Identity field to match. Supported values: email, subject."`
		MatchValue string `json:"matchValue" required:"true" doc:"Normalized email address or raw OIDC subject to match."`
	}
}

type DeleteAutoGrantRuleInput struct {
	RuleID string `path:"ruleId" required:"true" doc:"Platform auto-grant rule id."`
}

type AutoGrantRule struct {
	ID                 *uuid.UUID `json:"id,omitempty"`
	Role               string     `json:"role"`
	MatchType          string     `json:"matchType"`
	MatchValue         string     `json:"matchValue"`
	Source             string     `json:"source"`
	CreatedByAccountID *uuid.UUID `json:"createdByAccountId,omitempty"`
	CreatedAt          *time.Time `json:"createdAt,omitempty"`
	UpdatedAt          *time.Time `json:"updatedAt,omitempty"`
}

type AutoGrantRuleListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []AutoGrantRule `json:"items"`
	}
}

type AutoGrantRuleResponse struct {
	Status int           `json:"-"`
	Body   AutoGrantRule `json:"body"`
}
