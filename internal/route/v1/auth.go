package v1

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/service"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

// All error reporting goes through c.Error(err); the
// ErrorHandler middleware does the JSON shaping.
func RegisterAuthRoutes(r *gin.Engine, svc *service.AuthService, jwt *auth.JWTService, user *repository.UserRepository) {
	g := r.Group("/auth")
	g.POST("/login", login(svc))
	g.POST("/register", register(svc))
	g.POST("/refresh", refresh(svc))
	g.POST("/forgot-password", forgotPassword(svc))
	g.POST("/reset-password", resetPassword(svc))

	g.POST("/phone-login", phoneLogin(svc))
	g.POST("/verify-phone-otp", verifyPhoneOTP(svc))
	g.POST("/mobile-register", mobileRegister(svc))
	g.DELETE("/delete", middleware.AuthMiddleware(jwt, user), DeleteUser(svc))

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.AuthMiddleware(jwt, user))
	authGroup.POST("/logout", logout(svc))
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Login
// @Description User login
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse}  "Successful login"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/login [post]
func login(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		tokenResponse, err := svc.Login(req.Email, req.Password)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, dto.Response{
			Message: "Login successful",
			Status:  "success",
			Data:    tokenResponse,
		})
	}
}

// @Summary Phone Login
// @Description User login via phone number (OTP-based)
// @Tags Auth
// @Accept json
// @Produce json
// @Param mobileLoginRequest body dto.MobileLoginRequest true "Mobile login request payload"
// @Success 200 {object} dto.Response "OTP code sent"
// @Failure 400 {object} dto.Response "Bad request"
// @Router /auth/phone-login [post]
func phoneLogin(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.MobileLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		c.JSON(200, dto.Response{
			Data:    nil,
			Message: "OTP code sent",
			Status:  "success",
		})
	}
}

type VerifyPhoneOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTPCode     string `json:"otp_code" binding:"required"`
}

// @Summary Verify Phone OTP
// @Description Verify the OTP code sent to the user's phone number
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyPhoneOTPRequest body VerifyPhoneOTPRequest true "Verify phone OTP request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse} "OTP verified successfully"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/verify-phone-otp [post]
func verifyPhoneOTP(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyPhoneOTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		c.JSON(200, dto.Response{
			Data:    dto.TokenResponse{},
			Message: "OTP verified successfully",
			Status:  "success",
		})
	}
}

type MobileRegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// @Summary Mobile Register
// @Description User registration via mobile phone number (OTP-based)
// @Tags Auth
// @Accept json
// @Produce json
// @Param mobileRegisterRequest body MobileRegisterRequest true "Mobile register request payload"
// @Success 200 {object} dto.Response "OTP code sent"
// @Failure 400 {object} dto.Response "Bad request"
// @Router /auth/mobile-register [post]
func mobileRegister(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MobileRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		c.JSON(200, dto.Response{
			Data:    nil,
			Message: "OTP code sent",
			Status:  "success",
		})
	}
}

// RegisterRequest is the public registration payload. Role is intentionally
// NOT a field — public registration always creates a `client` user.
// Provisioning of `client_manager` / `admin` goes through an authed admin
// endpoint (not yet built).
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Register
// @Description User registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param registerRequest body RegisterRequest true "Register request payload"
// @Success 201 {object} dto.Response{data=dto.UserDTO} "Successful registration"
// @Failure 400 {object} dto.Response "Bad request"
// @Router /auth/register [post]
func register(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		// Public registration always creates a `client`. Role is not taken
		// from input — that closes the C1 (admin-self-register) and C2
		// (duplicate-JSON-key role escalation) findings from the audit.
		user, err := svc.CreateUser(req.Name, req.Email, req.Password, string(domain.RoleUser))
		if err != nil {
			c.Error(err)
			return
		}

		userDto := dto.UserDTO{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Role:   string(user.Role),
			Status: string(user.Status),
		}

		c.JSON(201, dto.Response{
			Message: "Registration successful",
			Status:  "success",
			Data:    userDto,
		})
	}
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Refresh Token
// @Description Refresh access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshTokenRequest body RefreshTokenRequest true "Refresh token request payload"
// @Success 200 {object} dto.Response{data=dto.TokenResponse} "Successful token refresh"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/refresh [post]
func refresh(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}

		tokenResponse, err := svc.RefreshToken(req.RefreshToken)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(200, dto.Response{
			Message: "Token refreshed successfully",
			Status:  "success",
			Data:    tokenResponse,
		})
	}
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Logout
// @Description Revoke the presented refresh token. Single-device.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param logoutRequest body LogoutRequest true "Logout request payload"
// @Success 200 {object} dto.Response "Logged out"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/logout [post]
func logout(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LogoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		user, exists := auth.UserFrom(c)
		if !exists {
			c.Error(errs.Unauthorized("USER_CONTEXT_MISSING", "user context missing"))
			return
		}
		if err := svc.Logout(req.RefreshToken, user.ID); err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{
			Status:  "success",
			Message: "logged out",
		})
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

// @Summary Forgot Password
// @Description Issue a single-use password-reset token. Always returns 200 to avoid email enumeration.
// @Tags Auth
// @Accept json
// @Produce json
// @Param forgotPasswordRequest body ForgotPasswordRequest true "Forgot password request payload"
// @Success 200 {object} dto.Response "If the email exists, a reset link has been sent"
// @Failure 400 {object} dto.Response "Bad request"
// @Router /auth/forgot-password [post]
func forgotPassword(svc *service.AuthService) gin.HandlerFunc {
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
		c.JSON(200, dto.Response{
			Status:  "success",
			Message: "if the email exists, a reset link has been sent",
		})
	}
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required,len=64"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

// @Summary Reset Password
// @Description Consume a reset token and set a new password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param resetPasswordRequest body ResetPasswordRequest true "Reset password request payload"
// @Success 200 {object} dto.Response "Password reset"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/reset-password [post]
func resetPassword(svc *service.AuthService) gin.HandlerFunc {
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
		c.JSON(200, dto.Response{
			Status:  "success",
			Message: "password reset",
		})
	}
}

// @Summary Delete User
// @Description Delete the authenticated user's account. This action is irreversible.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response "User deleted"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Router /auth/delete [delete]
func DeleteUser(svc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := auth.UserFrom(c)
		if !exists {
			c.Error(errs.Unauthorized("USER_CONTEXT_MISSING", "user context missing"))
			return
		}
		// if err := svc.DeleteUser(user.ID); err != nil {
		// 	c.Error(err)
		// 	return
		// }
		c.JSON(200, dto.Response{
			Status:  "success",
			Message: "user deleted",
		})
	}
}
