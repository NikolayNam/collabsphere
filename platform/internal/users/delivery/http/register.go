package http

import (
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler) {
	huma.Register(api, createUserOp, h.CreateUser)
}
