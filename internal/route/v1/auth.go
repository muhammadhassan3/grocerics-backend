package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/service"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthDeps struct {
	Admin *service.AdminAuthService
	Auth  *middleware.AuthDeps
}

func RegisterAuthRoutes(r *gin.Engine, d AuthDeps) {
	g := r.Group("/auth")

	// --- admin (web UI) ---
	g.POST("/login", adminLogin(d.Admin))
	g.POST("/refresh", adminRefresh(d.Admin))
	g.POST("/forgot-password", adminForgotPassword(d.Admin))
	g.POST("/reset-password", adminResetPassword(d.Admin))

	admin := r.Group("/auth")
	admin.Use(middleware.AuthMiddleware(d.Auth), middleware.AdminOnly())
	admin.POST("/logout", adminLogout(d.Admin))

	// --- client (mobile, OTP) ---
	g.POST("/phone-login", phoneLogin())
	g.POST("/verify-phone-otp", verifyPhoneOTP())
	g.POST("/mobile-register", mobileRegister())

	client := r.Group("/auth")
	client.Use(middleware.AuthMiddleware(d.Auth), middleware.ClientOnly())
	client.DELETE("/delete", deleteAccount())
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Admin login
// @Description Web-UI login with email + password. Clients use the phone/OTP routes.
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse}
// @Failure 401 {object} dto.Response
// @Router /auth/login [post]
func adminLogin(svc *service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		res, err := svc.Login(req.Email, req.Password)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Login successful", Data: res})
	}
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Refresh admin token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshTokenRequest body RefreshTokenRequest true "Refresh token request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse}
// @Failure 401 {object} dto.Response
// @Router /auth/refresh [post]
func adminRefresh(svc *service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		res, err := svc.RefreshToken(req.RefreshToken)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "Token refreshed", Data: res})
	}
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Admin logout
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param logoutRequest body LogoutRequest true "Logout request payload"
// @Success 200 {object} dto.Response
// @Router /auth/logout [post]
func adminLogout(svc *service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := svc.Logout(req.RefreshToken, auth.MustUser(c).ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "logged out"})
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

// @Summary Admin forgot password
// @Description Issue a single-use reset token. Always 200, to avoid email enumeration.
// @Tags Auth
// @Accept json
// @Produce json
// @Param forgotPasswordRequest body ForgotPasswordRequest true "Forgot password request payload"
// @Success 200 {object} dto.Response
// @Router /auth/forgot-password [post]
func adminForgotPassword(svc *service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := svc.ForgotPassword(req.Email); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "if the email exists, a reset link has been sent"})
	}
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required,len=64"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

// @Summary Admin reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param resetPasswordRequest body ResetPasswordRequest true "Reset password request payload"
// @Success 200 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Router /auth/reset-password [post]
func adminResetPassword(svc *service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		if err := svc.ResetPassword(req.Token, req.NewPassword); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "password reset"})
	}
}

// @Summary Phone login (client)
// @Description Send an OTP to the phone. STUB — OTP delivery not yet implemented.
// @Tags Auth
// @Accept json
// @Produce json
// @Param mobileLoginRequest body dto.MobileLoginRequest true "Mobile login request payload"
// @Success 200 {object} dto.Response
// @Router /auth/phone-login [post]
func phoneLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.MobileLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "OTP code sent"})
	}
}

type VerifyPhoneOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTPCode     string `json:"otp_code" binding:"required"`
}

// @Summary Verify phone OTP (client)
// @Description Verify the OTP and issue client tokens. STUB.
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyPhoneOTPRequest body VerifyPhoneOTPRequest true "Verify phone OTP request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse}
// @Router /auth/verify-phone-otp [post]
func verifyPhoneOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyPhoneOTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "OTP verified successfully", Data: dto.TokenResponse{}})
	}
}

type MobileRegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// @Summary Mobile register (client)
// @Description Register a client by phone (sends OTP). STUB.
// @Tags Auth
// @Accept json
// @Produce json
// @Param mobileRegisterRequest body MobileRegisterRequest true "Mobile register request payload"
// @Success 200 {object} dto.Response
// @Router /auth/mobile-register [post]
func mobileRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MobileRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "OTP code sent"})
	}
}

// @Summary Delete account (client)
// @Description Delete the authenticated client's account. STUB — deletion not yet wired.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /auth/delete [delete]
func deleteAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = auth.MustUser(c)
		c.JSON(200, dto.Response{Status: "success", Message: "user deleted"})
	}
}
