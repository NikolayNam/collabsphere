package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	create := createOrganizationOp
	create.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, create, h.CreateOrganization)

	huma.Register(api, getOrganizationByIdOp, h.GetOrganizationById)

	update := updateOrganizationOp
	update.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, update, h.UpdateOrganization)

	logoUpload := createOrganizationLogoUploadOp
	logoUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, logoUpload, h.CreateOrganizationLogoUpload)

	getCooperation := getCooperationApplicationOp
	getCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getCooperation, h.GetCooperationApplication)

	updateCooperation := updateCooperationApplicationOp
	updateCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateCooperation, h.UpdateCooperationApplication)

	submitCooperation := submitCooperationApplicationOp
	submitCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, submitCooperation, h.SubmitCooperationApplication)

	priceListUpload := createCooperationPriceListUploadOp
	priceListUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, priceListUpload, h.CreateCooperationPriceListUpload)

	legalUpload := createOrganizationLegalDocumentUploadOp
	legalUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, legalUpload, h.CreateOrganizationLegalDocumentUpload)

	createLegal := createOrganizationLegalDocumentOp
	createLegal.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, createLegal, h.CreateOrganizationLegalDocument)

	listLegal := listOrganizationLegalDocumentsOp
	listLegal.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listLegal, h.ListOrganizationLegalDocuments)
}
