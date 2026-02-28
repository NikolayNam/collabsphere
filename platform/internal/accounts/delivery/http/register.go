package http

import (
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler) {
	huma.Register(api, createAccountOp, h.CreateAccount)
	huma.Register(api, getAccountByIdOp, h.GetAccountById)
	huma.Register(api, getAccountByEmailOp, h.GetAccountByEmail)
}
