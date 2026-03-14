package http

import (
	"context"
	"net/http"
	"strings"

	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authdto "github.com/NikolayNam/collabsphere/internal/auth/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

type BrowserFlowConfig struct {
	DefaultReturnURL       string
	AllowedRedirectOrigins []string
	PublicBaseURL          string
}

type Handler struct {
	svc                  *authapp.Service
	passwordLoginEnabled bool
	zitadelAdminEnabled  bool
	browser              BrowserFlowConfig
	trustProxyHeaders    bool
}

func NewHandler(svc *authapp.Service, passwordLoginEnabled bool, zitadelAdminEnabled bool, browser BrowserFlowConfig, trustProxyHeaders bool) *Handler {
	return &Handler{
		svc:                  svc,
		passwordLoginEnabled: passwordLoginEnabled,
		zitadelAdminEnabled:  zitadelAdminEnabled,
		browser:              browser,
		trustProxyHeaders:    trustProxyHeaders,
	}
}

func (h *Handler) Login(ctx context.Context, input *authdto.LoginInput) (*authdto.TokenResponse, error) {
	if !h.passwordLoginEnabled {
		return nil, humaerr.From(ctx, fault.Forbidden("Password login is disabled. Use ZITADEL login.", fault.Code("PASSWORD_LOGIN_DISABLED")))
	}

	res, err := h.svc.Login(ctx, authapp.LoginCmd{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		UserAgent: optionalString(input.UserAgent),
		IP:        authmw.ExtractClientIP(input.XForwardedFor, input.XRealIP, "", authmw.ClientIPOptions{TrustProxyHeaders: h.trustProxyHeaders}),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	out := &authdto.TokenResponse{Status: http.StatusOK}
	out.Body.AccessToken = res.AccessToken
	out.Body.RefreshToken = res.RefreshToken
	out.Body.TokenType = res.TokenType
	out.Body.ExpiresIn = res.ExpiresIn
	return out, nil
}

func (h *Handler) ForceVerifyZitadelUserEmail(ctx context.Context, input *authdto.ForceVerifyZitadelUserEmailInput) (*authdto.ForceVerifyZitadelUserEmailResponse, error) {
	if err := requireAuthenticatedAccount(ctx); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if !h.zitadelAdminEnabled {
		return nil, humaerr.From(ctx, fault.Forbidden("ZITADEL admin email verification is disabled. Configure AUTH_ZITADEL_ADMIN_TOKEN to enable it.", fault.Code("AUTH_ZITADEL_ADMIN_DISABLED")))
	}

	res, err := h.svc.ForceVerifyZitadelUserEmail(ctx, authapp.ForceVerifyZitadelUserEmailCmd{UserID: input.UserID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	out := &authdto.ForceVerifyZitadelUserEmailResponse{Status: http.StatusOK}
	out.Body.UserID = res.UserID
	out.Body.Email = res.Email
	out.Body.Verified = res.Verified
	out.Body.AlreadyVerified = res.AlreadyVerified
	return out, nil
}

func (h *Handler) Exchange(ctx context.Context, input *authdto.ExchangeAuthTicketInput) (*authdto.TokenResponse, error) {
	res, err := h.svc.ExchangeAuthTicket(ctx, authapp.ExchangeAuthTicketCmd{
		Ticket:    input.Body.Ticket,
		UserAgent: optionalString(input.UserAgent),
		IP:        authmw.ExtractClientIP(input.XForwardedFor, input.XRealIP, "", authmw.ClientIPOptions{TrustProxyHeaders: h.trustProxyHeaders}),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	out := &authdto.TokenResponse{Status: http.StatusOK}
	out.Body.AccessToken = res.AccessToken
	out.Body.RefreshToken = res.RefreshToken
	out.Body.TokenType = res.TokenType
	out.Body.ExpiresIn = res.ExpiresIn
	out.Body.Provider = res.Provider
	out.Body.Intent = res.Intent
	out.Body.IsNewAccount = res.IsNewAccount
	return out, nil
}

func (h *Handler) Refresh(ctx context.Context, input *authdto.RefreshInput) (*authdto.TokenResponse, error) {
	res, err := h.svc.Refresh(ctx, authapp.RefreshCmd{
		RefreshToken: input.Body.RefreshToken,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	out := &authdto.TokenResponse{Status: http.StatusOK}
	out.Body.AccessToken = res.AccessToken
	out.Body.RefreshToken = res.RefreshToken
	out.Body.TokenType = res.TokenType
	out.Body.ExpiresIn = res.ExpiresIn
	return out, nil
}

func (h *Handler) Logout(ctx context.Context, input *authdto.LogoutInput) (*authdto.EmptyResponse, error) {
	if err := h.svc.Logout(ctx, authapp.LogoutCmd{
		RefreshToken: input.Body.RefreshToken,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &authdto.EmptyResponse{Status: http.StatusNoContent}, nil
}

func (h *Handler) Me(ctx context.Context, input *struct{}) (*authdto.MeResponse, error) {
	principal := authmw.PrincipalFromContext(ctx)

	acc, err := h.svc.Me(ctx, principal)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	resp := &authdto.MeResponse{Status: http.StatusOK}
	resp.Body.ID = acc.ID().UUID()
	resp.Body.Email = acc.Email().String()
	resp.Body.DisplayName = acc.DisplayName()
	resp.Body.AvatarObjectID = acc.AvatarObjectID()
	resp.Body.Bio = acc.Bio()
	resp.Body.Phone = acc.Phone()
	resp.Body.Locale = acc.Locale()
	resp.Body.Timezone = acc.Timezone()
	resp.Body.Website = acc.Website()
	resp.Body.IsActive = acc.IsActive()
	resp.Body.CreatedAt = acc.CreatedAt()
	resp.Body.UpdatedAt = acc.UpdatedAt()
	return resp, nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func requireAuthenticatedAccount(ctx context.Context) error {
	_, err := httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
	return err
}
