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
	svc                *application.Service
	localSignupEnabled bool
}

func NewHandler(svc *application.Service, localSignupEnabled bool) *Handler {
	return &Handler{svc: svc, localSignupEnabled: localSignupEnabled}
}

func (h *Handler) CreateAccount(ctx context.Context, input *dto.CreateAccountInput) (*dto.AccountResponse, error) {
	if !h.localSignupEnabled {
		return nil, humaerr.From(ctx, fault.Forbidden("Local signup is disabled. Use ZITADEL login.", fault.Code("LOCAL_SIGNUP_DISABLED")))
	}

	u, err := h.svc.CreateAccount(ctx, application.CreateAccountCmd{
		Email:       input.Body.Email,
		Password:    input.Body.Password,
		DisplayName: input.Body.DisplayName,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.accountResponse(ctx, u, http.StatusCreated)
}

func (h *Handler) GetAccountById(ctx context.Context, input *dto.GetAccountByIdInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountById(ctx, application.GetAccountByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.accountResponse(ctx, u, http.StatusOK)
}

func (h *Handler) GetAccountByEmail(ctx context.Context, input *dto.GetAccountByEmailInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountByEmail(ctx, application.GetAccountByEmailQuery{Email: input.Email})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.accountResponse(ctx, u, http.StatusOK)
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
	return h.accountProfileResponse(ctx, acc, http.StatusOK)
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
	return h.accountProfileResponse(ctx, acc, http.StatusOK)
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
	return h.accountProfileResponse(ctx, acc, http.StatusOK)
}

func (h *Handler) UploadMyVideo(ctx context.Context, input *dto.UploadMyVideoInput) (*dto.AccountProfileResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Account video file is required"))
	}
	defer form.File.Close()
	fileName := form.File.Filename
	if fileName == "" {
		fileName = "video.mp4"
	}
	if _, err := h.svc.UploadMyVideo(ctx, application.UploadMyVideoCmd{
		AccountID:   accountID,
		FileName:    fileName,
		ContentType: form.File.ContentType,
		SizeBytes:   form.File.Size,
		Body:        form.File,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	acc, err := h.svc.GetMyProfile(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.accountProfileResponse(ctx, acc, http.StatusOK)
}

func (h *Handler) ListMyVideos(ctx context.Context, _ *dto.ListMyVideosInput) (*dto.AccountVideosResponse, error) {
	accountID, err := principalAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListMyVideos(ctx, accountID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.AccountVideosResponse{Status: http.StatusOK}
	resp.Body.Items = make([]dto.AccountVideoItem, 0, len(items))
	for _, item := range items {
		resp.Body.Items = append(resp.Body.Items, dto.AccountVideoItem{
			ID:          item.ID,
			ObjectID:    item.ObjectID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			CreatedAt:   item.CreatedAt,
			SortOrder:   item.SortOrder,
		})
	}
	return resp, nil
}

func (h *Handler) accountResponse(ctx context.Context, a *accdomain.Account, status int) (*dto.AccountResponse, error) {
	resp := mapper.ToAccountResponse(a, status)
	if resp == nil || a == nil {
		return resp, nil
	}
	ids, err := h.svc.ListMyVideoObjectIDs(ctx, a.ID())
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp.Body.VideoObjectIDs = ids
	return resp, nil
}

func (h *Handler) accountProfileResponse(ctx context.Context, a *accdomain.Account, status int) (*dto.AccountProfileResponse, error) {
	resp := mapper.ToAccountProfileResponse(a, status)
	if resp == nil || a == nil {
		return resp, nil
	}
	ids, err := h.svc.ListMyVideoObjectIDs(ctx, a.ID())
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp.Body.VideoObjectIDs = ids
	return resp, nil
}

func principalAccountID(ctx context.Context) (accdomain.AccountID, error) {
	return httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
}
