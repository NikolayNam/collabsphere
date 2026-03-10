package dto

type OIDCBrowserStartInput struct {
	ReturnTo string `query:"return_to" doc:"Frontend callback URL or relative path that should receive the browser auth result." example:"http://localhost:3001/auth/callback"`
	Intent   string `query:"intent" default:"login" enum:"login,signup" doc:"Requested browser auth intent. Use signup to open the hosted registration flow."`
}

type OIDCBrowserSignupInput struct {
	ReturnTo string `query:"return_to" doc:"Frontend callback URL or relative path that should receive the browser auth result." example:"http://localhost:3001/auth/callback"`
}

type OIDCBrowserCallbackInput struct {
	UserAgent        string `header:"User-Agent"`
	XForwardedFor    string `header:"X-Forwarded-For"`
	XRealIP          string `header:"X-Real-IP"`
	State            string `query:"state" doc:"Opaque OIDC state returned by ZITADEL."`
	Code             string `query:"code" doc:"OIDC authorization code returned by ZITADEL on success."`
	Error            string `query:"error" doc:"OIDC provider error code returned by ZITADEL on failure."`
	ErrorDescription string `query:"error_description" doc:"Optional OIDC provider error description."`
}

type BrowserRedirectResponse struct {
	Status   int    `status:"303"`
	Location string `header:"Location"`
}
