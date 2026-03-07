package http

import (
	"context"
	"net/http"
	"strings"

	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authdto "github.com/NikolayNam/collabsphere/internal/auth/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

type Handler struct {
	svc *authapp.Service
}

func NewHandler(svc *authapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(ctx context.Context, input *authdto.LoginInput) (*authdto.TokenResponse, error) {
	res, err := h.svc.Login(ctx, authapp.LoginCmd{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		UserAgent: optionalString(input.UserAgent),
		IP:        extractClientIP(input.XForwardedFor, input.XRealIP),
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
	resp.Body.IsActive = acc.IsActive()
	return resp, nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func extractClientIP(forwardedFor, realIP string) *string {
	if value := firstHeaderValue(forwardedFor); value != "" {
		return &value
	}
	if value := strings.TrimSpace(realIP); value != "" {
		return &value
	}
	return nil
}

func firstHeaderValue(value string) string {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}
