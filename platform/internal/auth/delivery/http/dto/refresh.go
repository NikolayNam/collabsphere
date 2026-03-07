package dto

type RefreshInput struct {
	Body struct {
		RefreshToken string `json:"refreshToken" required:"true"`
	}
}
