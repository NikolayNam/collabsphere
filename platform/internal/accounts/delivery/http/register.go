package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	huma.Register(api, createAccountOp, h.CreateAccount)
	huma.Register(api, getAccountByIdOp, h.GetAccountById)
	huma.Register(api, getAccountByEmailOp, h.GetAccountByEmail)

	getMine := getMyAccountOp
	getMine.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getMine, h.GetMyAccount)

	updateMine := updateMyAccountOp
	updateMine.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateMine, h.UpdateMyAccount)

	avatarUpload := createAvatarUploadOp
	avatarUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, avatarUpload, h.CreateAvatarUpload)
}
