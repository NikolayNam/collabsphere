package dto

type ForceVerifyUserEmailInput struct {
	UserID string `path:"userId" required:"true" doc:"Raw ZITADEL user id whose email should be force-verified."`
}

type ForceVerifyUserEmailResponse struct {
	Status int `json:"-"`
	Body   struct {
		UserID          string `json:"userId"`
		Email           string `json:"email"`
		Verified        bool   `json:"verified"`
		AlreadyVerified bool   `json:"alreadyVerified"`
	}
}
