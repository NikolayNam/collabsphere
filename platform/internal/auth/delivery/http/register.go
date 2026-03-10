package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	huma.Register(api, loginOp, h.Login)
	huma.Register(api, zitadelLoginOp, h.BeginOIDCLogin)
	huma.Register(api, zitadelCallbackOp, h.CompleteOIDCCallback)
	huma.Register(api, refreshOp, h.Refresh)
	huma.Register(api, logoutOp, h.Logout)

	me := meOp
	me.Middlewares = huma.Middlewares{
		authmw.HumaAuthOptional(verifier),
	}
	huma.Register(api, me, h.Me)
}
