package service

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"
)

type ClientAuthService struct {
	users *repository.UserRepository
	jwt   *auth.JWTService

	mu    sync.Mutex
	codes map[string]otpEntry
}

type otpEntry struct {
	code      string
	expiresAt time.Time
}

const (
	otpTTL           = 5 * time.Minute
	clientAccessTTL  = 24 * time.Hour
	clientRefreshTTL = 30 * 24 * time.Hour
)

func NewClientAuthService(users *repository.UserRepository, jwt *auth.JWTService) *ClientAuthService {
	return &ClientAuthService{users: users, jwt: jwt, codes: map[string]otpEntry{}}
}

// RequestOTP generates a code for the phone, logs it (mock delivery), and
// returns it. TODO(dev-only): the real SMS-backed flow must NOT return the code
// — strip it from the handler response before production.
func (s *ClientAuthService) RequestOTP(phone string) (string, error) {
	code, err := generateOTP()
	if err != nil {
		return "", errs.Internal("OTP_GEN_FAILED", err)
	}
	s.mu.Lock()
	s.codes[phone] = otpEntry{code: code, expiresAt: time.Now().UTC().Add(otpTTL)}
	s.mu.Unlock()
	slog.Info("mock OTP issued", "phone", phone, "otp_code", code, "expires_in", otpTTL.String())
	return code, nil
}
func (s *ClientAuthService) VerifyOTP(phone, code string) (*dto.ClientAuthResponse, error) {
	s.mu.Lock()
	entry, ok := s.codes[phone]
	if ok && (entry.expiresAt.Before(time.Now().UTC()) || entry.code != code) {
		ok = false
	}
	if ok {
		delete(s.codes, phone)
	}
	s.mu.Unlock()
	if !ok {
		return nil, errs.Unauthorized("INVALID_OTP", "invalid or expired OTP")
	}

	u, err := s.users.FindByPhone(phone)
	if err != nil {
		return nil, errs.Internal("OTP_USER_LOOKUP_FAILED", err)
	}
	isNew := u == nil
	if isNew {
		u, err = s.users.Create(&domain.User{Name: phone, Phone: phone, Status: domain.UserStatusActive})
		if err != nil {
			return nil, errs.Internal("OTP_USER_CREATE_FAILED", err)
		}
	}

	access := s.jwt.Generate(u.Name, u.ID, string(domain.RoleUser), auth.KindClient, clientAccessTTL)
	refresh := s.jwt.Generate(u.Name, u.ID, string(domain.RoleUser), auth.KindClient, clientRefreshTTL)
	return &dto.ClientAuthResponse{
		TokenResponse: dto.TokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			UserData:     dto.UserData{ID: u.ID, Phone: u.Phone, FullName: u.Name, Role: string(domain.RoleUser)},
		},
		IsNew: isNew,
	}, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
