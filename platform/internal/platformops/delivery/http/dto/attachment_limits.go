package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListAttachmentLimitsInput struct {
	ScopeType string `query:"scopeType" doc:"Filter by scope type: platform, organization, account."`
	ScopeID  string `query:"scopeId" doc:"Filter by scope id (UUID)."`
}

type GetPlatformAttachmentLimitInput struct{}

type UpsertPlatformAttachmentLimitInput struct {
	Body struct {
		DocumentLimitBytes int64 `json:"documentLimitBytes" required:"true" doc:"Max size per document in bytes."`
		PhotoLimitBytes    int64 `json:"photoLimitBytes" required:"true" doc:"Max size per photo in bytes."`
		VideoLimitBytes    int64 `json:"videoLimitBytes" required:"true" doc:"Max size per video in bytes."`
		TotalLimitBytes    int64 `json:"totalLimitBytes" required:"true" doc:"Max total storage per user in bytes."`
	}
}

type GetOrganizationAttachmentLimitInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id."`
}

type UpsertOrganizationAttachmentLimitInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id."`
	Body          struct {
		DocumentLimitBytes int64 `json:"documentLimitBytes" required:"true" doc:"Max size per document in bytes."`
		PhotoLimitBytes    int64 `json:"photoLimitBytes" required:"true" doc:"Max size per photo in bytes."`
		VideoLimitBytes    int64 `json:"videoLimitBytes" required:"true" doc:"Max size per video in bytes."`
		TotalLimitBytes    int64 `json:"totalLimitBytes" required:"true" doc:"Max total storage per user in bytes."`
	}
}

type DeleteOrganizationAttachmentLimitInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id."`
}

type GetAccountAttachmentLimitInput struct {
	AccountID string `path:"accountId" required:"true" doc:"Account id."`
}

type UpsertAccountAttachmentLimitInput struct {
	AccountID string `path:"accountId" required:"true" doc:"Account id."`
	Body     struct {
		DocumentLimitBytes int64 `json:"documentLimitBytes" required:"true" doc:"Max size per document in bytes."`
		PhotoLimitBytes    int64 `json:"photoLimitBytes" required:"true" doc:"Max size per photo in bytes."`
		VideoLimitBytes    int64 `json:"videoLimitBytes" required:"true" doc:"Max size per video in bytes."`
		TotalLimitBytes    int64 `json:"totalLimitBytes" required:"true" doc:"Max total storage per user in bytes."`
	}
}

type DeleteAccountAttachmentLimitInput struct {
	AccountID string `path:"accountId" required:"true" doc:"Account id."`
}

type AttachmentLimit struct {
	ID                 uuid.UUID  `json:"id"`
	ScopeType          string     `json:"scopeType"`
	ScopeID            *uuid.UUID `json:"scopeId,omitempty"`
	DocumentLimitBytes int64      `json:"documentLimitBytes"`
	PhotoLimitBytes    int64      `json:"photoLimitBytes"`
	VideoLimitBytes    int64      `json:"videoLimitBytes"`
	TotalLimitBytes    int64      `json:"totalLimitBytes"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type AttachmentLimitListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []AttachmentLimit `json:"items"`
	}
}

type AttachmentLimitResponse struct {
	Status int             `json:"-"`
	Body   AttachmentLimit `json:"body"`
}
