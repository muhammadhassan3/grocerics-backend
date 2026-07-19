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

type AdminAuthService struct {
	admins           *repository.AdminRepository
	refreshTokens    *repository.AdminRefreshTokenRepository
	passwordResets   *repository.AdminPasswordResetRepository
	jwt              *auth.JWTService
	frontendResetURL string
}

func NewAdminAuthService(
	admins *repository.AdminRepository,
	refreshTokens *repository.AdminRefreshTokenRepository,
	passwordResets *repository.AdminPasswordResetRepository,
	jwt *auth.JWTService,
	frontendResetURL string,
) *AdminAuthService {
	return &AdminAuthService{
		admins:           admins,
		refreshTokens:    refreshTokens,
		passwordResets:   passwordResets,
		jwt:              jwt,
		frontendResetURL: frontendResetURL,
	}
}

const (
	adminAccessTTL  = 2 * time.Hour
	adminRefreshTTL = 7 * 24 * time.Hour
)

func (s *AdminAuthService) tokens(a *domain.Admin) (*dto.TokenResponse, error) {
	access := s.jwt.Generate(a.Name, a.ID, string(domain.RoleAdmin), auth.KindAdmin, adminAccessTTL)
	refresh := s.jwt.Generate(a.Name, a.ID, string(domain.RoleAdmin), auth.KindAdmin, adminRefreshTTL)
	if _, err := s.refreshTokens.Create(&domain.AdminRefreshToken{
		AdminID:   a.ID,
		TokenHash: util.HashToken(refresh),
		ExpiresAt: time.Now().UTC().Add(adminRefreshTTL),
	}); err != nil {
		return nil, errs.Internal("REFRESH_TOKEN_PERSIST_FAILED", err)
	}
	return &dto.TokenResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *AdminAuthService) Login(email, password string) (*dto.TokenResponse, error) {
	a, err := s.admins.FindByEmail(email)
	if err != nil {
		return nil, errs.Internal("LOGIN_LOOKUP_FAILED", err)
	}
	// Same error for unknown email and wrong password -> no info leak.
	if a == nil || !auth.VerifyPassword(a.PasswordHash, password) {
		return nil, errs.Unauthorized("INVALID_CREDENTIALS", "Invalid email or password")
	}
	return s.tokens(a)
}

func (s *AdminAuthService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	hash := util.HashToken(refreshToken)
	existing, err := s.refreshTokens.FindByTokenHash(hash)
	if err != nil {
		return nil, errs.Internal("REFRESH_LOOKUP_FAILED", err)
	}
	if existing == nil || !existing.IsActive(time.Now().UTC()) {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}
	claims, err := s.jwt.Validate(refreshToken)
	if err != nil {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token").WithCause(err)
	}
	a, err := s.admins.FindByID(claims.UserID)
	if err != nil {
		return nil, errs.Internal("REFRESH_USER_LOOKUP_FAILED", err)
	}
	if a == nil {
		return nil, errs.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}
	access := s.jwt.Generate(a.Name, a.ID, string(domain.RoleAdmin), auth.KindAdmin, adminAccessTTL)
	newRefresh := s.jwt.Generate(a.Name, a.ID, string(domain.RoleAdmin), auth.KindAdmin, adminRefreshTTL)
	existing.TokenHash = util.HashToken(newRefresh)
	existing.ExpiresAt = time.Now().UTC().Add(adminRefreshTTL)
	if _, err := s.refreshTokens.Update(existing); err != nil {
		return nil, errs.Internal("REFRESH_TOKEN_ROTATE_FAILED", err)
	}
	return &dto.TokenResponse{AccessToken: access, RefreshToken: newRefresh}, nil
}

func (s *AdminAuthService) Logout(refreshToken, adminID string) error {
	existing, err := s.refreshTokens.FindByTokenHashAndAdminID(util.HashToken(refreshToken), adminID)
	if err != nil {
		return errs.Internal("LOGOUT_LOOKUP_FAILED", err)
	}
	if existing == nil || !existing.IsActive(time.Now().UTC()) {
		return errs.Unauthorized("INVALID_REFRESH_TOKEN", "invalid refresh token")
	}
	if err := s.refreshTokens.Revoke(existing); err != nil {
		return errs.Internal("LOGOUT_REVOKE_FAILED", err)
	}
	return nil
}

func (s *AdminAuthService) ForgotPassword(email string) error {
	a, err := s.admins.FindByEmail(email)
	if err != nil {
		return errs.Internal("FORGOT_LOOKUP_FAILED", err)
	}
	if a == nil {
		return nil
	}
	raw, err := generateResetToken()
	if err != nil {
		return errs.Internal("FORGOT_TOKEN_GEN_FAILED", err)
	}
	reset := &domain.AdminPasswordReset{
		AdminID:   a.ID,
		TokenHash: util.HashToken(raw),
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
	if _, err := s.passwordResets.Create(reset); err != nil {
		return errs.Internal("FORGOT_PERSIST_FAILED", err)
	}
	// TODO: SMTP delivery. For now the token goes to logs
	slog.Info("admin password reset issued",
		"admin_id", a.ID, "email", a.Email,
		"expires_at", reset.ExpiresAt.Format(time.RFC3339),
		"reset_url", s.frontendResetURL+"?token="+raw,
		"reset_token", raw,
	)
	return nil
}

func (s *AdminAuthService) ResetPassword(rawToken, newPassword string) error {
	reset, err := s.passwordResets.FindByTokenHash(util.HashToken(rawToken))
	if err != nil {
		return errs.Internal("RESET_LOOKUP_FAILED", err)
	}
	now := time.Now().UTC()
	if reset == nil || reset.UsedAt != nil || reset.ExpiresAt.Before(now) {
		return errs.Unauthorized("INVALID_RESET_TOKEN", "invalid or expired reset token")
	}
	a, err := s.admins.FindByID(reset.AdminID)
	if err != nil {
		return errs.Internal("RESET_USER_LOOKUP_FAILED", err)
	}
	if a == nil {
		return errs.Unauthorized("INVALID_RESET_TOKEN", "invalid or expired reset token")
	}
	a.PasswordHash = auth.HashPassword(newPassword)
	if _, err := s.admins.Update(a); err != nil {
		return errs.Internal("RESET_UPDATE_FAILED", err)
	}
	if err := s.passwordResets.MarkUsed(reset); err != nil {
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
