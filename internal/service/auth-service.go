package service

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"
)

type AuthService struct {
	user             *repository.UserRepository
	refreshToken     *repository.RefreshTokenRepository
	passwordReset    *repository.PasswordResetRepository
	jwt              *auth.JWTService
	frontendResetURL string
}

func NewAuthService(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	passwordResetRepo *repository.PasswordResetRepository,
	jwt *auth.JWTService,
	frontendResetURL string,
) *AuthService {
	return &AuthService{
		user:             userRepo,
		refreshToken:     refreshTokenRepo,
		passwordReset:    passwordResetRepo,
		jwt:              jwt,
		frontendResetURL: frontendResetURL,
	}
}

func (a *AuthService) CreateUser(name, email, password, role string) (*domain.User, error) {
	r := domain.Role(role)
	if !r.IsValid() {
		r = domain.RoleUser
	}
	user := domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: auth.HashPassword(password),
		Role:         r,
		Status:       domain.UserStatusActive,
	}
	created, err := a.user.Create(&user)
	if err != nil {
		return nil, errs.BadRequest("CREATE_USER_FAILED", err.Error()).WithCause(err)
	}
	return created, nil
}

func (a *AuthService) Login(email, password string) (*dto.TokenResponse, error) {
	user, err := a.user.FindByEmail(email)
	if err != nil {
		return nil, errs.Internal("LOGIN_LOOKUP_FAILED", err)
	}
	// Same error for unknown email and wrong password -> no info leak.
	if user == nil || !auth.VerifyPassword(user.PasswordHash, password) {
		return nil, errs.Unauthorized("INVALID_CREDENTIALS", "Invalid email or password")
	}

	accessToken := a.jwt.Generate(user.Name, user.ID, string(user.Role), 2*time.Hour)
	refreshToken := a.jwt.Generate(user.Name, user.ID, string(user.Role), 7*24*time.Hour)

	if _, err = a.refreshToken.Create(&domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: util.HashToken(refreshToken),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}); err != nil {
		return nil, errs.Internal("REFRESH_TOKEN_PERSIST_FAILED", err)
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	hash := util.HashToken(refreshToken)

	existingToken, err := a.refreshToken.FindByTokenHash(hash)
	if err != nil {
		return nil, errs.Internal("REFRESH_LOOKUP_FAILED", err)
	}
	if existingToken == nil || !existingToken.IsActive(time.Now().UTC()) {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}

	claims, err := a.jwt.Validate(refreshToken)
	if err != nil {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token").WithCause(err)
	}

	user, err := a.user.FindByID(claims.UserID)
	if err != nil {
		return nil, errs.Internal("REFRESH_USER_LOOKUP_FAILED", err)
	}
	if user == nil {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}

	accessToken := a.jwt.Generate(user.Name, user.ID, string(user.Role), 2*time.Hour)
	newRefreshToken := a.jwt.Generate(user.Name, user.ID, string(user.Role), 7*24*time.Hour)

	existingToken.TokenHash = util.HashToken(newRefreshToken)
	existingToken.ExpiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)
	if _, err = a.refreshToken.Update(existingToken); err != nil {
		return nil, errs.Internal("REFRESH_TOKEN_ROTATE_FAILED", err)
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *AuthService) GetUserByEmail(email string) (*domain.User, error) {
	u, err := a.user.FindByEmail(email)
	if err != nil {
		return nil, errs.Internal("USER_LOOKUP_FAILED", err)
	}
	if u == nil {
		return nil, errs.NotFound("USER_NOT_FOUND", "user not found")
	}
	return u, nil
}

func (a *AuthService) GetUserByID(id string) (*domain.User, error) {
	u, err := a.user.FindByID(id)
	if err != nil {
		return nil, errs.Internal("USER_LOOKUP_FAILED", err)
	}
	if u == nil {
		return nil, errs.NotFound("USER_NOT_FOUND", "user not found")
	}
	return u, nil
}

func (a *AuthService) UpdateUser(id string, name string, password string, role string, email string) (*domain.User, error) {
	userData := domain.User{}
	userData.ID = id

	if name != "" {
		userData.Name = name
	}
	if password != "" {
		userData.PasswordHash = auth.HashPassword(password)
	}
	if role != "" {
		r := domain.Role(role)
		if r.IsValid() {
			userData.Role = r
		}
	}
	if email != "" {
		userData.Email = email
	}

	updated, err := a.user.Update(&userData)
	if err != nil {
		return nil, errs.BadRequest("UPDATE_USER_FAILED", err.Error()).WithCause(err)
	}
	return updated, nil
}

// Logout revokes the presented refresh token. Same generic error for
// unknown / already-revoked / expired so callers can't probe.
func (a *AuthService) Logout(refreshToken, userID string) error {
	hash := util.HashToken(refreshToken)
	existing, err := a.refreshToken.FindByRTokenHashAndUserID(hash, userID)
	if err != nil {
		return errs.Internal("LOGOUT_LOOKUP_FAILED", err)
	}
	if existing == nil || !existing.IsActive(time.Now().UTC()) {
		return errs.Unauthorized("INVALID_REFRESH_TOKEN", "invalid refresh token")
	}
	if _, err := a.refreshToken.Revoke(existing); err != nil {
		return errs.Internal("LOGOUT_REVOKE_FAILED", err)
	}
	return nil
}

// ForgotPassword issues a single-use reset token for the email if a
// user exists. Returns nil for both real and unknown emails so callers
// can't enumerate accounts.
func (a *AuthService) ForgotPassword(email string) error {
	user, err := a.user.FindByEmail(email)
	if err != nil {
		return errs.Internal("FORGOT_LOOKUP_FAILED", err)
	}
	if user == nil {
		return nil
	}
	rawToken, err := generateResetToken()
	if err != nil {
		return errs.Internal("FORGOT_TOKEN_GEN_FAILED", err)
	}
	reset := &domain.PasswordReset{
		UserID:    user.ID,
		TokenHash: util.HashToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
	if _, err := a.passwordReset.Create(reset); err != nil {
		return errs.Internal("FORGOT_PERSIST_FAILED", err)
	}
	// TODO: replace with SMTP delivery. For now the raw token goes to
	// server logs so dev/admin can forward it manually.
	slog.Info("password reset issued",
		"user_id", user.ID,
		"email", user.Email,
		"expires_at", reset.ExpiresAt.Format(time.RFC3339),
		"reset_url", a.frontendResetURL+"?token="+rawToken,
		"reset_token", rawToken,
	)
	return nil
}

// ResetPassword verifies the token, writes the new password, marks the
// reset row as used. Same generic error for every failure path.
func (a *AuthService) ResetPassword(rawToken, newPassword string) error {
	hash := util.HashToken(rawToken)
	reset, err := a.passwordReset.FindByTokenHash(hash)
	if err != nil {
		return errs.Internal("RESET_LOOKUP_FAILED", err)
	}
	now := time.Now().UTC()
	if reset == nil || reset.UsedAt != nil || reset.ExpiresAt.Before(now) {
		return errs.Unauthorized("INVALID_RESET_TOKEN", "invalid or expired reset token")
	}
	user, err := a.user.FindByID(reset.UserID)
	if err != nil {
		return errs.Internal("RESET_USER_LOOKUP_FAILED", err)
	}
	if user == nil {
		return errs.Unauthorized("INVALID_RESET_TOKEN", "invalid or expired reset token")
	}
	user.PasswordHash = auth.HashPassword(newPassword)
	if _, err := a.user.Update(user); err != nil {
		return errs.Internal("RESET_UPDATE_FAILED", err)
	}
	if err := a.passwordReset.MarkUsed(reset); err != nil {
		return errs.Internal("RESET_MARK_USED_FAILED", err)
	}
	return nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
