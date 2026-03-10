package ports

import "context"

type ZitadelUserEmailVerificationResult struct {
	UserID          string
	Email           string
	AlreadyVerified bool
}

type ZitadelAdminAPIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *ZitadelAdminAPIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	switch {
	case e.Code != "" && e.Message != "":
		return e.Code + ": " + e.Message
	case e.Message != "":
		return e.Message
	case e.Code != "":
		return e.Code
	default:
		return "zitadel admin api error"
	}
}

type ZitadelAdminClient interface {
	ForceVerifyUserEmail(ctx context.Context, userID string) (*ZitadelUserEmailVerificationResult, error)
}
