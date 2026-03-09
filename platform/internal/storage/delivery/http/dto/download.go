package dto

import (
	"time"

	"github.com/google/uuid"
)

type DownloadMyAvatarInput struct{}

type DownloadOrganizationLogoInput struct {
	OrganizationID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type DownloadCooperationPriceListInput struct {
	OrganizationID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type DownloadOrganizationLegalDocumentInput struct {
	OrganizationID string `path:"id" format:"uuid" doc:"Organization ID"`
	DocumentID     string `path:"document_id" format:"uuid" doc:"Organization legal document ID"`
}

type DownloadProductImportSourceInput struct {
	OrganizationID string `path:"organization_id" format:"uuid" doc:"Organization ID"`
	BatchID        string `path:"batch_id" format:"uuid" doc:"Product import batch ID"`
}

type DownloadChannelAttachmentInput struct {
	ChannelID string `path:"channel_id" format:"uuid" doc:"Channel ID"`
	ObjectID  string `path:"object_id" format:"uuid" doc:"Attachment object ID"`
}

type ListConferenceRecordingsInput struct {
	ConferenceID string `path:"conference_id" format:"uuid" doc:"Conference ID"`
}

type DownloadConferenceRecordingInput struct {
	ConferenceID string `path:"conference_id" format:"uuid" doc:"Conference ID"`
	RecordingID  string `path:"recording_id" format:"uuid" doc:"Conference recording ID"`
}

type ListMyFilesInput struct{}

type ListOrganizationFilesInput struct {
	OrganizationID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type FileItem struct {
	ObjectID       uuid.UUID  `json:"objectId"`
	OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
	FileName       string     `json:"fileName"`
	ContentType    *string    `json:"contentType,omitempty"`
	SizeBytes      int64      `json:"sizeBytes"`
	CreatedAt      time.Time  `json:"createdAt"`
	SourceType     string     `json:"sourceType" doc:"Logical source of the file in the system, for example account_avatar, organization_logo, product_image or legal_document."`
	SourceID       *uuid.UUID `json:"sourceId,omitempty" doc:"Identifier of the entity that references the file."`
}

type ConferenceRecordingItem struct {
	RecordingID  uuid.UUID  `json:"recordingId"`
	ConferenceID uuid.UUID  `json:"conferenceId"`
	ObjectID     uuid.UUID  `json:"objectId"`
	FileName     string     `json:"fileName"`
	ContentType  *string    `json:"contentType,omitempty"`
	SizeBytes    int64      `json:"sizeBytes"`
	CreatedAt    time.Time  `json:"createdAt"`
	CreatedBy    *uuid.UUID `json:"createdBy,omitempty"`
	DurationSec  *int32     `json:"durationSec,omitempty"`
	MimeType     *string    `json:"mimeType,omitempty"`
}

type ListFilesResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []FileItem `json:"items"`
	}
}

type ListConferenceRecordingsResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []ConferenceRecordingItem `json:"items"`
	}
}

type DownloadObjectResponse struct {
	Status int `json:"-"`
	Body   struct {
		ObjectID       uuid.UUID  `json:"objectId"`
		OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
		FileName       string     `json:"fileName"`
		ContentType    *string    `json:"contentType,omitempty"`
		SizeBytes      int64      `json:"sizeBytes"`
		DownloadURL    string     `json:"downloadUrl" doc:"Short-lived presigned URL for downloading the file bytes from S3-compatible storage."`
		ExpiresAt      time.Time  `json:"expiresAt"`
		CreatedAt      time.Time  `json:"createdAt"`
	}
}
