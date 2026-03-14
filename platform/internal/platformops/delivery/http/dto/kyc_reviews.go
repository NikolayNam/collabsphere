package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListKYCReviewsInput struct {
	Scope  string `query:"scope" doc:"Optional scope filter: account or organization."`
	Status string `query:"status" doc:"Optional KYC status filter: draft, submitted, in_review, needs_info, approved, rejected."`
	Limit  int    `query:"limit" doc:"Max items to return. Defaults to 50, capped at 200."`
	Offset int    `query:"offset" doc:"Pagination offset. Defaults to 0."`
}

type GetKYCReviewInput struct {
	ReviewID string `path:"reviewId" required:"true" doc:"Review id in form '<scope>:<uuid>'."`
}

type DecideKYCReviewInput struct {
	ReviewID string `path:"reviewId" required:"true" doc:"Review id in form '<scope>:<uuid>'."`
	Body     struct {
		Decision string  `json:"decision" required:"true" doc:"Decision: approve, reject, request_info."`
		Reason   *string `json:"reason,omitempty" doc:"Optional reviewer comment/reason."`
	}
}

type DecideKYCDocumentInput struct {
	ReviewID   string `path:"reviewId" required:"true" doc:"Review id in form '<scope>:<uuid>'."`
	DocumentID string `path:"documentId" required:"true" doc:"Document id UUID within selected KYC review scope."`
	Body       struct {
		Decision string  `json:"decision" required:"true" doc:"Decision: approve, reject, request_info."`
		Reason   *string `json:"reason,omitempty" doc:"Optional reviewer comment/reason."`
	}
}

type ListKYCLevelsInput struct {
	Scope string `query:"scope" doc:"Optional scope filter: account or organization."`
}

type CreateKYCLevelInput struct {
	Body struct {
		Scope                 string                `json:"scope" required:"true" doc:"Scope: account or organization."`
		Code                  string                `json:"code" required:"true" doc:"Stable level code, unique within scope."`
		Name                  string                `json:"name" required:"true" doc:"Human-readable level name."`
		Rank                  int                   `json:"rank" required:"true" doc:"Priority rank, higher means stricter level."`
		IsActive              bool                  `json:"isActive" doc:"Whether level is active for assignment."`
		RequiredDocumentTypes []KYCLevelRequirement `json:"requiredDocumentTypes" doc:"Document requirements for this level."`
	}
}

type UpdateKYCLevelInput struct {
	LevelID string `path:"levelId" required:"true" doc:"KYC level UUID for update."`
	Body    struct {
		Scope                 string                `json:"scope" required:"true" doc:"Scope: account or organization."`
		Code                  string                `json:"code" required:"true" doc:"Stable level code, unique within scope."`
		Name                  string                `json:"name" required:"true" doc:"Human-readable level name."`
		Rank                  int                   `json:"rank" required:"true" doc:"Priority rank, higher means stricter level."`
		IsActive              bool                  `json:"isActive" doc:"Whether level is active for assignment."`
		RequiredDocumentTypes []KYCLevelRequirement `json:"requiredDocumentTypes" doc:"Document requirements for this level."`
	}
}

type DeleteKYCLevelInput struct {
	LevelID string `path:"levelId" required:"true" doc:"KYC level UUID."`
}

type IssueKYCLevelInput struct {
	ReviewID string `path:"reviewId" required:"true" doc:"Review id in form '<scope>:<uuid>'."`
}

type KYCReviewListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Total int             `json:"total"`
		Items []KYCReviewItem `json:"items"`
	}
}

type KYCReviewResponse struct {
	Status int           `json:"-"`
	Body   KYCReviewItem `json:"body"`
}

type KYCLevelListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []KYCLevel `json:"items"`
	}
}

type KYCLevelResponse struct {
	Status int      `json:"-"`
	Body   KYCLevel `json:"body"`
}

type IssueKYCLevelResponse struct {
	Status int `json:"-"`
	Body   struct {
		LevelID   *uuid.UUID `json:"levelId,omitempty"`
		LevelCode *string    `json:"levelCode,omitempty"`
		LevelName *string    `json:"levelName,omitempty"`
		IssuedAt  *time.Time `json:"issuedAt,omitempty"`
	}
}

type KYCReviewItem struct {
	ReviewID           string                  `json:"reviewId"`
	Scope              string                  `json:"scope"`
	SubjectID          uuid.UUID               `json:"subjectId"`
	Status             string                  `json:"status"`
	KYCLevelCode       *string                 `json:"kycLevelCode,omitempty"`
	KYCLevelName       *string                 `json:"kycLevelName,omitempty"`
	LegalName          *string                 `json:"legalName,omitempty"`
	CountryCode        *string                 `json:"countryCode,omitempty"`
	RegistrationNumber *string                 `json:"registrationNumber,omitempty"`
	TaxID              *string                 `json:"taxId,omitempty"`
	DocumentNumber     *string                 `json:"documentNumber,omitempty"`
	ResidenceAddress   *string                 `json:"residenceAddress,omitempty"`
	ReviewNote         *string                 `json:"reviewNote,omitempty"`
	ReviewerAccountID  *uuid.UUID              `json:"reviewerAccountId,omitempty"`
	SubmittedAt        *time.Time              `json:"submittedAt,omitempty"`
	ReviewedAt         *time.Time              `json:"reviewedAt,omitempty"`
	CreatedAt          *time.Time              `json:"createdAt,omitempty"`
	UpdatedAt          time.Time               `json:"updatedAt"`
	Documents          []KYCDocumentReviewItem `json:"documents,omitempty"`
	Events             []KYCReviewEvent        `json:"events,omitempty"`
}

type KYCDocumentReviewItem struct {
	ID                uuid.UUID  `json:"id"`
	ObjectID          uuid.UUID  `json:"objectId"`
	DocumentType      string     `json:"documentType"`
	Title             string     `json:"title"`
	Status            string     `json:"status"`
	ReviewNote        *string    `json:"reviewNote,omitempty"`
	ReviewerAccountID *uuid.UUID `json:"reviewerAccountId,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
}

type KYCReviewEvent struct {
	ID                uuid.UUID `json:"id"`
	Scope             string    `json:"scope"`
	SubjectID         uuid.UUID `json:"subjectId"`
	Decision          string    `json:"decision"`
	Reason            *string   `json:"reason,omitempty"`
	ReviewerAccountID uuid.UUID `json:"reviewerAccountId"`
	CreatedAt         time.Time `json:"createdAt"`
}

type KYCLevel struct {
	ID                    uuid.UUID             `json:"id"`
	Scope                 string                `json:"scope"`
	Code                  string                `json:"code"`
	Name                  string                `json:"name"`
	Rank                  int                   `json:"rank"`
	IsActive              bool                  `json:"isActive"`
	RequiredDocumentTypes []KYCLevelRequirement `json:"requiredDocumentTypes"`
	CreatedAt             time.Time             `json:"createdAt"`
	UpdatedAt             time.Time             `json:"updatedAt"`
}

type KYCLevelRequirement struct {
	DocumentType string `json:"documentType"`
	MinCount     int    `json:"minCount"`
}
