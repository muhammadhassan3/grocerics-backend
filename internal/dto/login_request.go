package dto

// @Swagger:model LoginRequest
// @Property email: Account email address
// @Property password: Account password
// @Description Credentials submitted to the login endpoint.
type LoginRequest struct {
	// Account email address
	Email string `json:"email" binding:"required,email"`
	// Account password
	Password string `json:"password" binding:"required"`
}

// @Swagger:model MobileLoginRequest
// @Property phone_number: Phone number to send the login OTP to, in E.164 format
// @Description Request payload submitted to the mobile app's phone-based login endpoint.
type MobileLoginRequest struct {
	// Phone number to send the login OTP to, in E.164 format
	PhoneNumber string `json:"phone_number" binding:"required"`
}
