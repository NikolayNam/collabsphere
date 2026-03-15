package dto

type UpdateProductInput struct {
	OrganizationID string `path:"organization_id"`
	ProductID      string `path:"product_id"`
	Body           struct {
		CategoryID   *string `json:"categoryId,omitempty" format:"uuid"`
		Status       *string `json:"status,omitempty" maxLength:"24"`
		Name         *string `json:"name,omitempty" maxLength:"255"`
		Description  *string `json:"description,omitempty"`
		SKU          *string `json:"sku,omitempty" maxLength:"128"`
		PriceAmount  *string `json:"priceAmount,omitempty" example:"199.90"`
		CurrencyCode *string `json:"currencyCode,omitempty" minLength:"0" maxLength:"3"`
		IsActive     *bool   `json:"isActive,omitempty"`
	}
}
