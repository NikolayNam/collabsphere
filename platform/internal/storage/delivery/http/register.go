package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	op := downloadMyAvatarOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadMyAvatar)

	op = downloadMyAccountVideoOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadMyAccountVideo)

	op = downloadOrganizationLogoOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadOrganizationLogo)

	op = downloadOrganizationVideoOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadOrganizationVideo)

	op = downloadCooperationPriceListOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadCooperationPriceList)

	op = downloadOrganizationLegalDocumentOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadOrganizationLegalDocument)

	op = downloadProductImportSourceOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadProductImportSource)

	op = downloadProductVideoOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadProductVideo)

	op = downloadChannelAttachmentOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadChannelAttachment)

	op = listConferenceRecordingsOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.ListConferenceRecordings)

	op = downloadConferenceRecordingOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.DownloadConferenceRecording)

	op = listMyFilesOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.ListMyFiles)

	op = listOrganizationFilesOp
	op.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, op, h.ListOrganizationFiles)
}
