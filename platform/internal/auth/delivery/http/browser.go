package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authdto "github.com/NikolayNam/collabsphere/internal/auth/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

func (h *Handler) BeginOIDCBrowserLogin(ctx context.Context, input *authdto.OIDCBrowserStartInput) (*authdto.BrowserRedirectResponse, error) {
	intent := normalizeBrowserIntent(input.Intent)
	if intent == "" {
		return nil, humaerr.From(ctx, fault.Validation("invalid OIDC intent"))
	}
	redirectTarget, err := h.beginBrowserRedirect(ctx, input.ReturnTo, intent)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &authdto.BrowserRedirectResponse{
		Status:   http.StatusSeeOther,
		Location: redirectTarget,
	}, nil
}

func (h *Handler) BeginOIDCBrowserSignup(ctx context.Context, input *authdto.OIDCBrowserSignupInput) (*authdto.BrowserRedirectResponse, error) {
	redirectTarget, err := h.beginBrowserRedirect(ctx, input.ReturnTo, "signup")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &authdto.BrowserRedirectResponse{
		Status:   http.StatusSeeOther,
		Location: redirectTarget,
	}, nil
}

func (h *Handler) beginBrowserRedirect(ctx context.Context, rawReturnTo, intent string) (string, error) {
	returnTo, err := h.resolveBrowserReturnTo(rawReturnTo)
	if err != nil {
		return "", err
	}

	res, err := h.svc.BeginOIDCLogin(ctx, authapp.BeginOIDCLoginCmd{
		ReturnTo: returnTo,
		Intent:   intent,
	})
	if err != nil {
		return "", err
	}
	return res.AuthorizationURL, nil
}

func (h *Handler) CompleteOIDCBrowserCallback(ctx context.Context, input *authdto.OIDCBrowserCallbackInput) (*authdto.BrowserRedirectResponse, error) {
	state := strings.TrimSpace(input.State)
	returnTo := h.lookupCallbackReturnTo(ctx, state)

	if providerErr := strings.TrimSpace(input.Error); providerErr != "" {
		redirectTarget, ok := h.redirectWithBrowserError(returnTo, providerErr, strings.TrimSpace(input.ErrorDescription))
		if ok {
			return &authdto.BrowserRedirectResponse{
				Status:   http.StatusSeeOther,
				Location: redirectTarget,
			}, nil
		}
		return nil, humaerr.From(ctx, fault.Unauthorized(providerErr))
	}

	res, err := h.svc.CompleteOIDCCallback(ctx, authapp.CompleteOIDCCallbackCmd{
		State:     state,
		Code:      strings.TrimSpace(input.Code),
		UserAgent: optionalString(input.UserAgent),
		IP:        authmw.ExtractClientIP(input.XForwardedFor, input.XRealIP, "", authmw.ClientIPOptions{TrustProxyHeaders: h.trustProxyHeaders}),
	})
	if err != nil {
		code, description := browserErrorPayload(err)
		redirectTarget, ok := h.redirectWithBrowserError(returnTo, code, description)
		if ok {
			return &authdto.BrowserRedirectResponse{
				Status:   http.StatusSeeOther,
				Location: redirectTarget,
			}, nil
		}
		return nil, humaerr.From(ctx, err)
	}

	redirectTarget, err := addQueryParams(res.ReturnTo, map[string]string{"ticket": res.ExchangeTicket})
	if err != nil {
		return nil, humaerr.From(ctx, fault.Internal("failed to build callback redirect", fault.WithCause(err)))
	}
	return &authdto.BrowserRedirectResponse{
		Status:   http.StatusSeeOther,
		Location: redirectTarget,
	}, nil
}

func (h *Handler) lookupCallbackReturnTo(ctx context.Context, state string) string {
	if strings.TrimSpace(state) == "" {
		return h.normalizeBrowserReturnTarget(strings.TrimSpace(h.browser.DefaultReturnURL))
	}
	res, err := h.svc.ResolveOIDCCallbackState(ctx, authapp.ResolveOIDCCallbackStateCmd{State: state})
	if err != nil || res == nil {
		return h.normalizeBrowserReturnTarget(strings.TrimSpace(h.browser.DefaultReturnURL))
	}
	return h.normalizeBrowserReturnTarget(strings.TrimSpace(res.ReturnTo))
}

func (h *Handler) resolveBrowserReturnTo(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(h.browser.DefaultReturnURL)
	}
	if value == "" {
		return "", fault.Validation("browser return URL is required")
	}
	if strings.HasPrefix(value, "/") {
		return h.normalizeBrowserReturnTarget(value), nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fault.Validation("browser return URL is invalid")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fault.Validation("browser return URL must be absolute or start with '/'")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fault.Validation("browser return URL scheme is invalid")
	}
	origin := parsed.Scheme + "://" + parsed.Host
	for _, allowed := range h.allowedBrowserOrigins() {
		if origin == allowed {
			return parsed.String(), nil
		}
	}
	return "", fault.Forbidden("browser return URL origin is not allowed")
}

func (h *Handler) allowedBrowserOrigins() []string {
	origins := make([]string, 0, len(h.browser.AllowedRedirectOrigins)+2)
	seen := make(map[string]struct{}, len(h.browser.AllowedRedirectOrigins)+2)
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			value = parsed.Scheme + "://" + parsed.Host
		} else {
			value = strings.TrimRight(value, "/")
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		origins = append(origins, value)
	}
	for _, origin := range h.browser.AllowedRedirectOrigins {
		add(origin)
	}
	if parsed, err := url.Parse(strings.TrimSpace(h.browser.DefaultReturnURL)); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		add(parsed.Scheme + "://" + parsed.Host)
	}
	if parsed, err := url.Parse(strings.TrimSpace(h.browser.PublicBaseURL)); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		add(parsed.Scheme + "://" + parsed.Host)
	}
	return origins
}

func (h *Handler) normalizeBrowserReturnTarget(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") {
		return value
	}
	base := strings.TrimSpace(h.browser.PublicBaseURL)
	if base == "" {
		return value
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return value
	}
	return parsed.Scheme + "://" + parsed.Host + value
}

func (h *Handler) redirectWithBrowserError(returnTo, code, description string) (string, bool) {
	returnTo = strings.TrimSpace(returnTo)
	if returnTo == "" {
		return "", false
	}
	params := map[string]string{"error": code}
	if strings.TrimSpace(description) != "" {
		params["error_description"] = description
	}
	redirectTarget, err := addQueryParams(returnTo, params)
	if err != nil {
		return "", false
	}
	return redirectTarget, true
}

func addQueryParams(target string, values map[string]string) (string, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	for key, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func browserErrorPayload(err error) (string, string) {
	if appErr, ok := fault.As(err); ok && appErr != nil {
		switch appErr.Kind {
		case fault.KindValidation:
			return "invalid_request", appErr.Message
		case fault.KindUnauthorized:
			return "unauthorized", appErr.Message
		case fault.KindForbidden:
			return "access_denied", appErr.Message
		case fault.KindUnavailable:
			return "temporarily_unavailable", appErr.Message
		default:
			return "server_error", "Authentication failed"
		}
	}
	return "server_error", "Authentication failed"
}

func normalizeBrowserIntent(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "login":
		return "login"
	case "signup":
		return "signup"
	default:
		return ""
	}
}
