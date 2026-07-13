package dto

// @Swagger:model PhoneOTPResponse
// @Property token: Opaque token identifying the OTP challenge, required to verify the code
// @Property expires_at: Expiration timestamp of the OTP, RFC3339
// @Description Response returned after an OTP has been sent to a phone number.
type PhoneOTPResponse struct {
	// Opaque token identifying the OTP challenge, required to verify the code
	Token string `json:"token"`
	// Expiration timestamp of the OTP, RFC3339
	ExpiresAt string `json:"expires_at"`
}
