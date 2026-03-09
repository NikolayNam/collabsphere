package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	op := downloadObjectOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadObject)

	op = listMyFilesOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.ListMyFiles)

	op = listOrganizationFilesOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.ListOrganizationFiles)
}
