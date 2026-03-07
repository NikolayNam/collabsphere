package dto

type LogoutInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken" required:"true"`
	}
}
