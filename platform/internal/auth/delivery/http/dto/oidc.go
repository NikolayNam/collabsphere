package dto

type OIDCLoginResponse struct {
	Status int `json:"-"`
	Body   struct {
		AuthorizationURL string `json:"authorizationUrl"`
	}
}

type OIDCCallbackInput struct {
	UserAgent        string `header:"User-Agent"`
	XForwardedFor    string `header:"X-Forwarded-For"`
	XRealIP          string `header:"X-Real-IP"`
	State            string `query:"state" required:"true"`
	Code             string `query:"code" required:"true"`
	Error            string `query:"error"`
	ErrorDescription string `query:"error_description"`
}
