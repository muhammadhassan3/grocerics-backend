package dto

// @Swagger:model TokenResponse
// @Property access_token: JWT access token used to authenticate subsequent requests
// @Property refresh_token: Token used to obtain a new access token once the current one expires
// @Description Response structure for authentication tokens.
type TokenResponse struct {
	// JWT access token used to authenticate subsequent requests
	AccessToken string `json:"access_token"`
	// Token used to obtain a new access token once the current one expires
	RefreshToken string `json:"refresh_token"`
}
