package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	createAccount := createAccountOp
	if !h.localSignupEnabled {
		createAccount.Description += "\n\n> [!CAUTION]\n> This legacy local-signup endpoint is disabled in this environment. Use `GET /auth/zitadel/signup` instead."
	}
	huma.Register(api, createAccount, h.CreateAccount)
	huma.Register(api, getAccountByIdOp, h.GetAccountById)
	huma.Register(api, getAccountByEmailOp, h.GetAccountByEmail)

	getMine := getMyAccountOp
	getMine.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getMine, h.GetMyAccount)

	updateMine := updateMyAccountOp
	updateMine.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateMine, h.UpdateMyAccount)

	uploadAvatar := uploadMyAvatarOp
	uploadAvatar.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadAvatar, h.UploadMyAvatar)

	uploadVideo := uploadMyVideoOp
	uploadVideo.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadVideo, h.UploadMyVideo)

	listVideos := listMyVideosOp
	listVideos.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listVideos, h.ListMyVideos)

	getKYC := getMyKYCOp
	getKYC.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getKYC, h.GetMyKYC)

	updateKYC := updateMyKYCOp
	updateKYC.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateKYC, h.UpdateMyKYC)

	createKYCDocUpload := createMyKYCDocumentUploadOp
	createKYCDocUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, createKYCDocUpload, h.CreateMyKYCDocumentUpload)

	completeKYCDocUpload := completeMyKYCDocumentUploadOp
	completeKYCDocUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, completeKYCDocUpload, h.CompleteMyKYCDocumentUpload)
}
