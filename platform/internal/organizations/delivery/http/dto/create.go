package dto

type CreateOrganizationInput struct {
	Body struct {
		LegalName    string  `json:"legalName" required:"true" example:"ООО Рога и Копыты" maxProperties:"200"`
		DisplayName  *string `json:"displayName,omitempty"`
		PrimaryEmail string  `json:"primaryEmail" required:"true" format:"email"`
	}
}
