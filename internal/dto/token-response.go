package dto

// @Swagger:model UserData
// @Property image_url: URL of the user's profile image
// @Property name: Name of the user (same field as GET/PATCH /v1/me)
// @Property role: Role of the user
// @Description User data associated with the authenticated user.
type UserData struct {
	ID       string `json:"id,omitempty"`
	Phone    string `json:"phone,omitempty"`
	ImageURL string `json:"image_url"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type ClientAuthResponse struct {
	TokenResponse
	IsNew bool `json:"is_new"`
}

// @Swagger:model TokenResponse
// @Property access_token: JWT access token used to authenticate subsequent requests
// @Property refresh_token: Token used to obtain a new access token once the current one expires
// @Description Response structure for authentication tokens.
type TokenResponse struct {
	// JWT access token used to authenticate subsequent requests
	AccessToken string `json:"access_token"`
	// Token used to obtain a new access token once the current one expires
	RefreshToken string `json:"refresh_token"`
	// User data associated with the authenticated user
	UserData UserData `json:"user_data"`
}
