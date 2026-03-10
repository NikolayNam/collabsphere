package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationBody struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	Slug           string      `json:"slug"`
	LogoObjectID   *uuid.UUID  `json:"logoObjectId,omitempty"`
	VideoObjectIDs []uuid.UUID `json:"videoObjectIds,omitempty"`
	Description    *string     `json:"description,omitempty"`
	Website        *string     `json:"website,omitempty"`
	PrimaryEmail   *string     `json:"primaryEmail,omitempty"`
	Phone          *string     `json:"phone,omitempty"`
	Address        *string     `json:"address,omitempty"`
	Industry       *string     `json:"industry,omitempty"`
	IsActive       bool        `json:"isActive"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      *time.Time  `json:"updatedAt,omitempty"`
}

type OrganizationResponse struct {
	Status int              `json:"-"`
	Body   OrganizationBody `json:"body"`
}

type OrganizationVideosResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []OrganizationVideoItem `json:"items"`
	} `json:"body"`
}

type OrganizationVideoResponse struct {
	Status int                   `json:"-"`
	Body   OrganizationVideoItem `json:"body"`
}

type OrganizationVideoItem struct {
	ID          uuid.UUID  `json:"id"`
	ObjectID    uuid.UUID  `json:"objectId"`
	FileName    string     `json:"fileName"`
	ContentType *string    `json:"contentType,omitempty"`
	SizeBytes   int64      `json:"sizeBytes"`
	CreatedAt   time.Time  `json:"createdAt"`
	UploadedBy  *uuid.UUID `json:"uploadedBy,omitempty"`
	SortOrder   int64      `json:"sortOrder"`
}

type MemberBody struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"created_at"`
}

type MembersResponse struct {
	OrganizationID uuid.UUID  `json:"organizationId"`
	Status         int        `json:"-"`
	Body           MemberBody `json:"body"`
}

type MembersListBody struct {
	Data []MemberBody `json:"data"`
}

type MembersListResponse struct {
	Status int             `json:"-"`
	Body   MembersListBody `json:"-"`
}
