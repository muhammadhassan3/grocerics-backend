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
