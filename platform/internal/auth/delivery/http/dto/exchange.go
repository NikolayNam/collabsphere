package dto

type ExchangeAuthTicketInput struct {
	UserAgent     string `header:"User-Agent"`
	XForwardedFor string `header:"X-Forwarded-For"`
	XRealIP       string `header:"X-Real-IP"`

	Body struct {
		Ticket string `json:"ticket" required:"true"`
	}
}
