package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

func Register(router chi.Router, api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	_ = router

	huma.Register(api, zitadelLoginOp, h.BeginOIDCBrowserLogin)
	huma.Register(api, zitadelSignupOp, h.BeginOIDCBrowserSignup)
	huma.Register(api, zitadelCallbackOp, h.CompleteOIDCBrowserCallback)

	login := loginOp
	if !h.passwordLoginEnabled {
		login.Description += "\n\n> [!CAUTION]\n> This legacy password-login endpoint is disabled in this environment. Use `GET /auth/zitadel/login` to sign in or `GET /auth/zitadel/signup` to register."
	}
	huma.Register(api, login, h.Login)

	huma.Register(api, exchangeOp, h.Exchange)
	huma.Register(api, refreshOp, h.Refresh)
	huma.Register(api, logoutOp, h.Logout)

	me := meOp
	me.Middlewares = huma.Middlewares{
		authmw.HumaAuthOptional(verifier),
	}
	huma.Register(api, me, h.Me)
}
