package dto

type ListOrganizationsInput struct {
	Limit int `query:"limit" default:"100" minimum:"1" maximum:"500"`
}

type OrganizationListItemBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ListOrganizationsResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []OrganizationListItemBody `json:"items"`
	} `json:"body"`
}
