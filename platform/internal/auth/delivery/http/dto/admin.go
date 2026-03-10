package dto

type ForceVerifyZitadelUserEmailInput struct {
	UserID string `path:"userId" required:"true" doc:"ZITADEL user id whose email should be force-verified."`
}

type ForceVerifyZitadelUserEmailResponse struct {
	Status int `json:"-"`
	Body   struct {
		UserID          string `json:"userId"`
		Email           string `json:"email"`
		Verified        bool   `json:"verified"`
		AlreadyVerified bool   `json:"alreadyVerified"`
	}
}
