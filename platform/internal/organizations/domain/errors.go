package domain

import "errors"

var (
	ErrNowRequired       = errors.New("now is required")
	ErrTimestampsMissing = errors.New("timestamps are required")
	ErrTimestampsInvalid = errors.New("timestamps are invalid")

	ErrOrganizationIDEmpty = errors.New("organization id is empty")

	ErrEmailEmpty   = errors.New("email is empty")
	ErrEmailInvalid = errors.New("email is invalid")
	ErrEmailTooLong = errors.New("email is too long")

	ErrNameInvalid         = errors.New("name is invalid")
	ErrSlugInvalid         = errors.New("slug is invalid")
	ErrDescriptionInvalid  = errors.New("description is invalid")
	ErrWebsiteInvalid      = errors.New("website is invalid")
	ErrPrimaryEmailInvalid = errors.New("primary email is invalid")
	ErrPhoneInvalid        = errors.New("phone is invalid")
	ErrAddressInvalid      = errors.New("address is invalid")
	ErrIndustryInvalid     = errors.New("industry is invalid")

	ErrInvalidOrganizationStatus = errors.New("invalid organization status")

	ErrMembershipInvalid = errors.New("membership is invalid")

	ErrCooperationApplicationIDEmpty           = errors.New("cooperation application id is empty")
	ErrCooperationApplicationStatusInvalid     = errors.New("cooperation application status is invalid")
	ErrCooperationConfirmationEmailInvalid     = errors.New("cooperation confirmation email is invalid")
	ErrCooperationCompanyNameInvalid           = errors.New("cooperation company name is invalid")
	ErrCooperationRepresentedCategoriesInvalid = errors.New("cooperation represented categories is invalid")
	ErrCooperationMinimumOrderAmountInvalid    = errors.New("cooperation minimum order amount is invalid")
	ErrCooperationDeliveryGeographyInvalid     = errors.New("cooperation delivery geography is invalid")
	ErrCooperationSalesChannelsInvalid         = errors.New("cooperation sales channels is invalid")
	ErrCooperationStorefrontURLInvalid         = errors.New("cooperation storefront url is invalid")
	ErrCooperationContactFirstNameInvalid      = errors.New("cooperation contact first name is invalid")
	ErrCooperationContactLastNameInvalid       = errors.New("cooperation contact last name is invalid")
	ErrCooperationContactJobTitleInvalid       = errors.New("cooperation contact job title is invalid")
	ErrCooperationContactEmailInvalid          = errors.New("cooperation contact email is invalid")
	ErrCooperationContactPhoneInvalid          = errors.New("cooperation contact phone is invalid")
	ErrCooperationPartnerCodeInvalid           = errors.New("cooperation partner code is invalid")
	ErrCooperationReviewNoteInvalid            = errors.New("cooperation review note is invalid")
	ErrCooperationApplicationIncomplete        = errors.New("cooperation application is incomplete")

	ErrOrganizationLegalDocumentIDEmpty           = errors.New("organization legal document id is empty")
	ErrOrganizationLegalDocumentObjectIDEmpty     = errors.New("organization legal document object id is empty")
	ErrOrganizationLegalDocumentTypeInvalid       = errors.New("organization legal document type is invalid")
	ErrOrganizationLegalDocumentTitleInvalid      = errors.New("organization legal document title is invalid")
	ErrOrganizationLegalDocumentStatusInvalid     = errors.New("organization legal document status is invalid")
	ErrOrganizationLegalDocumentReviewNoteInvalid = errors.New("organization legal document review note is invalid")
)
