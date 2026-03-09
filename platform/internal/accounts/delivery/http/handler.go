package http

import (
	"context"
	"net/http"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"

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

func (h *Handler) UploadMyAvatar(ctx context.Context, input *dto.UploadMyAvatarInput) (*dto.AccountProfileResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Avatar file is required"))
	}
	defer form.File.Close()

	fileName := form.File.Filename
	if fileName == "" {
		fileName = "avatar.bin"
	}

	acc, err := h.svc.UploadAvatar(ctx, application.UploadAvatarCmd{
		AccountID:   accountID,
		FileName:    fileName,
		ContentType: form.File.ContentType,
		SizeBytes:   form.File.Size,
		Body:        form.File,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountProfileResponse(acc, http.StatusOK), nil
}

func principalAccountID(ctx context.Context) (accdomain.AccountID, error) {
	return httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
}
