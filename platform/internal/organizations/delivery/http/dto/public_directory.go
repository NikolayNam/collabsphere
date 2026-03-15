package dto

type ListPublicKYCDirectoryInput struct {
	Limit int `query:"limit" default:"100" minimum:"1" maximum:"500"`
}

type PublicKYCDirectoryOrganizationBody struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description,omitempty"`
	Website       *string `json:"website,omitempty"`
	Industry      *string `json:"industry,omitempty"`
	PrimaryDomain *string `json:"primaryDomain,omitempty"`
	KYCLevelCode  string  `json:"kycLevelCode"`
	KYCLevelName  *string `json:"kycLevelName,omitempty"`
}

type PublicKYCDirectoryResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []PublicKYCDirectoryOrganizationBody `json:"items"`
	} `json:"body"`
}
