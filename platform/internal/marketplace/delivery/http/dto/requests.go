package dto

type ListOrdersInput struct {
	Limit  int `query:"limit" default:"50" minimum:"1" maximum:"200"`
	Offset int `query:"offset" default:"0" minimum:"0"`
}

type CreateOrderInput struct {
	Body struct {
		OrganizationID string  `json:"organizationId" required:"true" format:"uuid"`
		Title          string  `json:"title" required:"true" maxLength:"255"`
		Description    *string `json:"description,omitempty"`
		BudgetAmount   *string `json:"budgetAmount,omitempty"`
		CurrencyCode   *string `json:"currencyCode,omitempty" minLength:"3" maxLength:"3"`
		Comment        *string `json:"comment,omitempty"`
		Items          []struct {
			CategoryID  *string `json:"categoryId,omitempty" format:"uuid"`
			ProductName *string `json:"productName,omitempty" maxLength:"255"`
			Quantity    *string `json:"quantity,omitempty"`
			Unit        *string `json:"unit,omitempty" maxLength:"32"`
			Note        *string `json:"note,omitempty"`
		} `json:"items,omitempty"`
	}
}

type GetOrderInput struct {
	OrderID string `path:"order_id" format:"uuid"`
}

type AddOrderCommentInput struct {
	OrderID string `path:"order_id" format:"uuid"`
	Body    struct {
		OrganizationID string `json:"organizationId" required:"true" format:"uuid"`
		Comment        string `json:"comment" required:"true"`
	}
}

type CreateOfferInput struct {
	OrderID string `path:"order_id" format:"uuid"`
	Body    struct {
		OrganizationID string  `json:"organizationId" required:"true" format:"uuid"`
		Comment        *string `json:"comment,omitempty"`
		Items          []struct {
			CategoryID   *string `json:"categoryId,omitempty" format:"uuid"`
			ProductID    *string `json:"productId,omitempty" format:"uuid"`
			CustomTitle  *string `json:"customTitle,omitempty" maxLength:"255"`
			Quantity     *string `json:"quantity,omitempty"`
			Unit         *string `json:"unit,omitempty" maxLength:"32"`
			PriceAmount  *string `json:"priceAmount,omitempty"`
			CurrencyCode *string `json:"currencyCode,omitempty" minLength:"3" maxLength:"3"`
			Note         *string `json:"note,omitempty"`
		} `json:"items,omitempty"`
	}
}

type ListOffersInput struct {
	OrderID string `path:"order_id" format:"uuid"`
}

type AddOfferCommentInput struct {
	OfferID string `path:"offer_id" format:"uuid"`
	Body    struct {
		OrganizationID string `json:"organizationId" required:"true" format:"uuid"`
		Comment        string `json:"comment" required:"true"`
	}
}
