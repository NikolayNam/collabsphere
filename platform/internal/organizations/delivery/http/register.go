package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	create := createOrganizationOp
	create.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, create, h.CreateOrganization)

	listMine := listMyOrganizationsOp
	listMine.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listMine, h.ListMyOrganizations)

	huma.Register(api, getOrganizationByIdOp, h.GetOrganizationById)
	huma.Register(api, resolveOrganizationByHostOp, h.ResolveOrganizationByHost)

	update := updateOrganizationOp
	update.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, update, h.UpdateOrganization)

	uploadLogo := uploadOrganizationLogoOp
	uploadLogo.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadLogo, h.UploadOrganizationLogo)

	uploadVideo := uploadOrganizationVideoOp
	uploadVideo.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadVideo, h.UploadOrganizationVideo)

	listVideos := listOrganizationVideosOp
	listVideos.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listVideos, h.ListOrganizationVideos)

	getCooperation := getCooperationApplicationOp
	getCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getCooperation, h.GetCooperationApplication)

	updateCooperation := updateCooperationApplicationOp
	updateCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateCooperation, h.UpdateCooperationApplication)

	submitCooperation := submitCooperationApplicationOp
	submitCooperation.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, submitCooperation, h.SubmitCooperationApplication)

	uploadPriceList := uploadCooperationPriceListOp
	uploadPriceList.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadPriceList, h.UploadCooperationPriceList)

	createLegalDocumentUpload := createOrganizationLegalDocumentUploadOp
	createLegalDocumentUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, createLegalDocumentUpload, h.CreateOrganizationLegalDocumentUpload)

	completeLegalDocumentUpload := completeOrganizationLegalDocumentUploadOp
	completeLegalDocumentUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, completeLegalDocumentUpload, h.CompleteOrganizationLegalDocumentUpload)

	getKYCProfile := getOrganizationKYCProfileOp
	getKYCProfile.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getKYCProfile, h.GetOrganizationKYCProfile)

	updateKYCProfile := updateOrganizationKYCProfileOp
	updateKYCProfile.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, updateKYCProfile, h.UpdateOrganizationKYCProfile)

	createKYCDocumentUpload := createOrganizationKYCDocumentUploadOp
	createKYCDocumentUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, createKYCDocumentUpload, h.CreateOrganizationKYCDocumentUpload)

	completeKYCDocumentUpload := completeOrganizationKYCDocumentUploadOp
	completeKYCDocumentUpload.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, completeKYCDocumentUpload, h.CompleteOrganizationKYCDocumentUpload)

	uploadLegalDocument := uploadOrganizationLegalDocumentOp
	uploadLegalDocument.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, uploadLegalDocument, h.UploadOrganizationLegalDocument)

	listLegalDocuments := listOrganizationLegalDocumentsOp
	listLegalDocuments.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, listLegalDocuments, h.ListOrganizationLegalDocuments)

	getKYCRequirements := getOrganizationKYCRequirementsOp
	getKYCRequirements.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getKYCRequirements, h.GetOrganizationKYCRequirements)

	getAnalysis := getOrganizationLegalDocumentAnalysisOp
	getAnalysis.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getAnalysis, h.GetOrganizationLegalDocumentAnalysis)

	getVerification := getOrganizationLegalDocumentVerificationOp
	getVerification.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, getVerification, h.GetOrganizationLegalDocumentVerification)

	reprocessAnalysis := reprocessOrganizationLegalDocumentAnalysisOp
	reprocessAnalysis.Middlewares = huma.Middlewares{authmw.HumaAuthOptional(verifier)}
	huma.Register(api, reprocessAnalysis, h.ReprocessOrganizationLegalDocumentAnalysis)
}
