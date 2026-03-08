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

	legalDocumentUpload := createOrganizationLegalDocumentUploadOp
	legalDocumentUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, legalDocumentUpload, h.CreateOrganizationLegalDocumentUpload)

	createLegalDocument := createOrganizationLegalDocumentOp
	createLegalDocument.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, createLegalDocument, h.CreateOrganizationLegalDocument)

	listLegalDocuments := listOrganizationLegalDocumentsOp
	listLegalDocuments.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listLegalDocuments, h.ListOrganizationLegalDocuments)

	getAnalysis := getOrganizationLegalDocumentAnalysisOp
	getAnalysis.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getAnalysis, h.GetOrganizationLegalDocumentAnalysis)

	reprocessAnalysis := reprocessOrganizationLegalDocumentAnalysisOp
	reprocessAnalysis.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, reprocessAnalysis, h.ReprocessOrganizationLegalDocumentAnalysis)
}
