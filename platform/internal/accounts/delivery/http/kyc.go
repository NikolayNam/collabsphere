package http

import (
	"context"
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

func (h *Handler) GetMyKYC(ctx context.Context, _ *dto.GetMyKYCInput) (*dto.AccountKYCResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	profile, docs, err := h.svc.GetMyKYCProfile(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AccountKYCResponse{Status: http.StatusOK}
	out.Body.AccountID = profile.AccountID
	out.Body.Status = profile.Status
	out.Body.LegalName = profile.LegalName
	out.Body.CountryCode = profile.CountryCode
	out.Body.DocumentNumber = profile.DocumentNumber
	out.Body.ResidenceAddress = profile.ResidenceAddress
	out.Body.ReviewNote = profile.ReviewNote
	out.Body.ReviewerAccount = profile.ReviewerAccount
	out.Body.SubmittedAt = profile.SubmittedAt
	out.Body.ReviewedAt = profile.ReviewedAt
	out.Body.CreatedAt = profile.CreatedAt
	out.Body.UpdatedAt = profile.UpdatedAt
	out.Body.Documents = make([]dto.AccountKYCDocument, 0, len(docs))
	for _, item := range docs {
		out.Body.Documents = append(out.Body.Documents, toAccountKYCDocumentDTO(item))
	}
	return out, nil
}

func (h *Handler) UpdateMyKYC(ctx context.Context, input *dto.UpdateMyKYCInput) (*dto.AccountKYCResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	updated, err := h.svc.UpdateMyKYCProfile(ctx, application.UpdateMyKYCProfileCmd{
		AccountID:        accountID,
		Status:           input.Body.Status,
		LegalName:        input.Body.LegalName,
		CountryCode:      input.Body.CountryCode,
		DocumentNumber:   input.Body.DocumentNumber,
		ResidenceAddress: input.Body.ResidenceAddress,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AccountKYCResponse{Status: http.StatusOK}
	out.Body.AccountID = updated.AccountID
	out.Body.Status = updated.Status
	out.Body.LegalName = updated.LegalName
	out.Body.CountryCode = updated.CountryCode
	out.Body.DocumentNumber = updated.DocumentNumber
	out.Body.ResidenceAddress = updated.ResidenceAddress
	out.Body.ReviewNote = updated.ReviewNote
	out.Body.ReviewerAccount = updated.ReviewerAccount
	out.Body.SubmittedAt = updated.SubmittedAt
	out.Body.ReviewedAt = updated.ReviewedAt
	out.Body.CreatedAt = updated.CreatedAt
	out.Body.UpdatedAt = updated.UpdatedAt
	out.Body.Documents = []dto.AccountKYCDocument{}
	return out, nil
}

func (h *Handler) CreateMyKYCDocumentUpload(ctx context.Context, input *dto.CreateMyKYCDocumentUploadInput) (*dto.AccountKYCUploadResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateMyKYCDocumentUpload(ctx, application.CreateMyKYCDocumentUploadCmd{
		AccountID:      accountID,
		DocumentType:   input.Body.DocumentType,
		Title:          input.Body.Title,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AccountKYCUploadResponse{Status: http.StatusCreated}
	out.Body.ID = result.UploadID
	out.Body.OrganizationID = nil
	out.Body.ObjectID = result.ObjectID
	out.Body.CreatedByAccountID = accountID.UUID()
	out.Body.Purpose = "account_kyc_document"
	out.Body.Status = "pending"
	out.Body.Bucket = result.Bucket
	out.Body.ObjectKey = result.ObjectKey
	out.Body.FileName = result.FileName
	out.Body.ContentType = input.Body.ContentType
	out.Body.DeclaredSizeBytes = result.SizeBytes
	out.Body.ChecksumSHA256 = input.Body.ChecksumSHA256
	out.Body.Metadata = map[string]any{
		"documentType": result.DocumentType,
		"title":        result.Title,
	}
	out.Body.UploadURL = &result.UploadURL
	out.Body.ExpiresAt = &result.ExpiresAt
	return out, nil
}

func (h *Handler) CompleteMyKYCDocumentUpload(ctx context.Context, input *dto.CompleteMyKYCDocumentUploadInput) (*dto.AccountKYCDocumentResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	uploadID, err := httpbind.ParseUUID(input.UploadID, fault.Validation("Upload id is invalid", fault.Field("upload_id", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	document, err := h.svc.CompleteMyKYCDocumentUpload(ctx, application.CompleteMyKYCDocumentUploadCmd{
		AccountID: accountID,
		UploadID:  uploadID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.AccountKYCDocumentResponse{
		Status: http.StatusOK,
		Body:   toAccountKYCDocumentDTO(*document),
	}, nil
}

func toAccountKYCDocumentDTO(item application.AccountKYCDocumentView) dto.AccountKYCDocument {
	return dto.AccountKYCDocument{
		ID:              item.ID,
		AccountID:       item.AccountID,
		ObjectID:        item.ObjectID,
		DocumentType:    item.DocumentType,
		Title:           item.Title,
		Status:          item.Status,
		ReviewNote:      item.ReviewNote,
		ReviewerAccount: item.ReviewerAccount,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		ReviewedAt:      item.ReviewedAt,
	}
}
