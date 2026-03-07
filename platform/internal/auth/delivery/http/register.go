package http

import "github.com/danielgtaylor/huma/v2"

func Register(api huma.API, h *Handler) {
	huma.Register(api, loginOp, h.Login)
	huma.Register(api, refreshOp, h.Refresh)
	huma.Register(api, logoutOp, h.Logout)
	huma.Register(api, meOp, h.Me)
}
