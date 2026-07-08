package dto

// @Swagger:model TokenResponse
// @Description: Response structure for authentication tokens
// @Property access_token: The JWT access token for authentication
// @Property refresh_token: The token used to obtain a new access token when the current one expires
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
