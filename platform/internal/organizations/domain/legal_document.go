package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type OrganizationLegalDocumentStatus string

const (
	OrganizationLegalDocumentStatusPending  OrganizationLegalDocumentStatus = "pending"
	OrganizationLegalDocumentStatusApproved OrganizationLegalDocumentStatus = "approved"
	OrganizationLegalDocumentStatusRejected OrganizationLegalDocumentStatus = "rejected"
)

type OrganizationLegalDocument struct {
	id                  uuid.UUID
	organizationID      OrganizationID
	documentType        string
	status              OrganizationLegalDocumentStatus
	objectID            uuid.UUID
	title               string
	uploadedByAccountID *uuid.UUID
	reviewerAccountID   *uuid.UUID
	reviewNote          *string
	createdAt           time.Time
	updatedAt           *time.Time
	reviewedAt          *time.Time
	deletedAt           *time.Time
}

type NewOrganizationLegalDocumentParams struct {
	ID                  uuid.UUID
	OrganizationID      OrganizationID
	DocumentType        string
	ObjectID            uuid.UUID
	Title               string
	UploadedByAccountID *uuid.UUID
	Now                 time.Time
}

func NewOrganizationLegalDocument(p NewOrganizationLegalDocumentParams) (*OrganizationLegalDocument, error) {
	if p.ID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.ObjectID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentObjectIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}
	documentType, err := normalizeLegalDocumentType(p.DocumentType)
	if err != nil {
		return nil, err
	}
	title, err := normalizeLegalDocumentTitle(p.Title)
	if err != nil {
		return nil, err
	}
	updatedAt := p.Now
	return &OrganizationLegalDocument{
		id:                  p.ID,
		organizationID:      p.OrganizationID,
		documentType:        documentType,
		status:              OrganizationLegalDocumentStatusPending,
		objectID:            p.ObjectID,
		title:               title,
		uploadedByAccountID: cloneUUIDPtr(p.UploadedByAccountID),
		createdAt:           p.Now,
		updatedAt:           &updatedAt,
	}, nil
}

type RehydrateOrganizationLegalDocumentParams struct {
	ID                  uuid.UUID
	OrganizationID      OrganizationID
	DocumentType        string
	Status              string
	ObjectID            uuid.UUID
	Title               string
	UploadedByAccountID *uuid.UUID
	ReviewerAccountID   *uuid.UUID
	ReviewNote          *string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	ReviewedAt          *time.Time
	DeletedAt           *time.Time
}

func RehydrateOrganizationLegalDocument(p RehydrateOrganizationLegalDocumentParams) (*OrganizationLegalDocument, error) {
	if p.ID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.ObjectID == uuid.Nil {
		return nil, ErrOrganizationLegalDocumentObjectIDEmpty
	}
	if p.CreatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	documentType, err := normalizeLegalDocumentType(p.DocumentType)
	if err != nil {
		return nil, err
	}
	status, err := normalizeLegalDocumentStatus(p.Status)
	if err != nil {
		return nil, err
	}
	title, err := normalizeLegalDocumentTitle(p.Title)
	if err != nil {
		return nil, err
	}
	reviewNote, err := normalizeOptionalOrgField(p.ReviewNote, 4096, ErrOrganizationLegalDocumentReviewNoteInvalid)
	if err != nil {
		return nil, err
	}
	return &OrganizationLegalDocument{
		id:                  p.ID,
		organizationID:      p.OrganizationID,
		documentType:        documentType,
		status:              status,
		objectID:            p.ObjectID,
		title:               title,
		uploadedByAccountID: cloneUUIDPtr(p.UploadedByAccountID),
		reviewerAccountID:   cloneUUIDPtr(p.ReviewerAccountID),
		reviewNote:          reviewNote,
		createdAt:           p.CreatedAt,
		updatedAt:           cloneTimePtr(p.UpdatedAt),
		reviewedAt:          cloneTimePtr(p.ReviewedAt),
		deletedAt:           cloneTimePtr(p.DeletedAt),
	}, nil
}

func (d *OrganizationLegalDocument) ID() uuid.UUID                           { return d.id }
func (d *OrganizationLegalDocument) OrganizationID() OrganizationID          { return d.organizationID }
func (d *OrganizationLegalDocument) DocumentType() string                    { return d.documentType }
func (d *OrganizationLegalDocument) Status() OrganizationLegalDocumentStatus { return d.status }
func (d *OrganizationLegalDocument) ObjectID() uuid.UUID                     { return d.objectID }
func (d *OrganizationLegalDocument) Title() string                           { return d.title }
func (d *OrganizationLegalDocument) UploadedByAccountID() *uuid.UUID {
	return cloneUUIDPtr(d.uploadedByAccountID)
}
func (d *OrganizationLegalDocument) ReviewerAccountID() *uuid.UUID {
	return cloneUUIDPtr(d.reviewerAccountID)
}
func (d *OrganizationLegalDocument) ReviewNote() *string    { return cloneStringPtr(d.reviewNote) }
func (d *OrganizationLegalDocument) CreatedAt() time.Time   { return d.createdAt }
func (d *OrganizationLegalDocument) UpdatedAt() *time.Time  { return cloneTimePtr(d.updatedAt) }
func (d *OrganizationLegalDocument) ReviewedAt() *time.Time { return cloneTimePtr(d.reviewedAt) }
func (d *OrganizationLegalDocument) DeletedAt() *time.Time  { return cloneTimePtr(d.deletedAt) }

func normalizeLegalDocumentType(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
		return "", ErrOrganizationLegalDocumentTypeInvalid
	}
	return trimmed, nil
}

func normalizeLegalDocumentTitle(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 255 {
		return "", ErrOrganizationLegalDocumentTitleInvalid
	}
	return trimmed, nil
}

func normalizeLegalDocumentStatus(value string) (OrganizationLegalDocumentStatus, error) {
	switch OrganizationLegalDocumentStatus(strings.TrimSpace(value)) {
	case OrganizationLegalDocumentStatusPending,
		OrganizationLegalDocumentStatusApproved,
		OrganizationLegalDocumentStatusRejected:
		return OrganizationLegalDocumentStatus(strings.TrimSpace(value)), nil
	default:
		return "", ErrOrganizationLegalDocumentStatusInvalid
	}
}
