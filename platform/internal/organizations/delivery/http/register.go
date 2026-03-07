package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	create := createOrganizationOp
	create.Middlewares = huma.Middlewares{
		authmw.HumaAuthOptional(verifier),
	}

	huma.Register(api, create, h.CreateOrganization)
	huma.Register(api, getOrganizationByIdOp, h.GetOrganizationById)
}
