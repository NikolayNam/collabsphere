package http

import "github.com/danielgtaylor/huma/v2"

func Register(api huma.API, h *Handler, verifier AccessTokenVerifier) {
	huma.Register(api, loginOp, h.Login)
	huma.Register(api, refreshOp, h.Refresh)
	huma.Register(api, logoutOp, h.Logout)

	me := meOp
	me.Middlewares = huma.Middlewares{
		authPrincipalMiddleware(verifier),
	}
	huma.Register(api, me, h.Me)
}
