package http

import (
	"context"
	"net/http"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/mapper"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateAccount(ctx context.Context, input *dto.CreateAccountInput) (*dto.AccountResponse, error) {
	u, err := h.svc.CreateAccount(ctx, application.CreateAccountCmd{
		Email:       input.Body.Email,
		Password:    input.Body.Password,
		DisplayName: input.Body.DisplayName,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	return mapper.ToAccountResponse(u, http.StatusCreated), nil
}

func (h *Handler) GetAccountById(ctx context.Context, input *dto.GetAccountByIdInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountById(ctx, application.GetAccountByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountResponse(u, http.StatusOK), nil
}

func (h *Handler) GetAccountByEmail(ctx context.Context, input *dto.GetAccountByEmailInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountByEmail(ctx, application.GetAccountByEmailQuery{Email: input.Email})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountResponse(u, http.StatusOK), nil
}

func (h *Handler) GetMyAccount(ctx context.Context, _ *dto.GetMyAccountInput) (*dto.AccountProfileResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	acc, err := h.svc.GetMyProfile(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountProfileResponse(acc, http.StatusOK), nil
}

func (h *Handler) UpdateMyAccount(ctx context.Context, input *dto.UpdateMyAccountProfileInput) (*dto.AccountProfileResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	acc, err := h.svc.UpdateMyProfile(ctx, application.UpdateMyProfileCmd{
		AccountID:      accountID,
		DisplayName:    input.Body.DisplayName,
		AvatarObjectID: input.Body.AvatarObjectID,
		ClearAvatar:    input.Body.ClearAvatar,
		Bio:            input.Body.Bio,
		Phone:          input.Body.Phone,
		Locale:         input.Body.Locale,
		Timezone:       input.Body.Timezone,
		Website:        input.Body.Website,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountProfileResponse(acc, http.StatusOK), nil
}

func (h *Handler) CreateAvatarUpload(ctx context.Context, input *dto.CreateAvatarUploadInput) (*dto.UploadResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateAvatarUpload(ctx, application.CreateAvatarUploadCmd{
		AccountID:      accountID,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.UploadResponse{Status: http.StatusCreated}
	resp.Body.ObjectID = result.ObjectID
	resp.Body.Bucket = result.Bucket
	resp.Body.ObjectKey = result.ObjectKey
	resp.Body.UploadURL = result.UploadURL
	resp.Body.ExpiresAt = result.ExpiresAt
	resp.Body.FileName = result.FileName
	resp.Body.SizeBytes = result.SizeBytes
	return resp, nil
}

func principalAccountID(ctx context.Context) (accdomain.AccountID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return accdomain.AccountID{}, fault.Unauthorized("Authentication required")
	}
	accountID, err := accdomain.AccountIDFromUUID(principal.AccountID)
	if err != nil {
		return accdomain.AccountID{}, fault.Unauthorized("Authentication required")
	}
	return accountID, nil
}
