package repository

import (
	"context"
	"time"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *domain.RefreshToken) (*domain.RefreshToken, error) {
	err := gorm.G[domain.RefreshToken](r.db).Create(context.Background(), token)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return token, nil
}

func (r *RefreshTokenRepository) Update(token *domain.RefreshToken) (*domain.RefreshToken, error) {
	_, err := gorm.G[domain.RefreshToken](r.db).
		Where("id = ?", token.ID).
		Updates(context.Background(), *token)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return token, nil
}

// FindByTokenHash looks up by the SHA-256 of the raw token. Callers must
// hash before calling — never pass the raw JWT here.
func (r *RefreshTokenRepository) FindByTokenHash(hash string) (*domain.RefreshToken, error) {
	data, err := gorm.G[domain.RefreshToken](r.db).
		Where("token_hash = ?", hash).
		First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return &data, nil
}

func (r *RefreshTokenRepository) FindByRTokenHashAndUserID(hash, userID string) (*domain.RefreshToken, error) {
	data, err := gorm.G[domain.RefreshToken](r.db).
		Where("token_hash = ? AND user_id = ?", hash, userID).
		First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return &data, nil
}

func (r *RefreshTokenRepository) Revoke(token *domain.RefreshToken) (*domain.RefreshToken, error) {
	now := time.Now().UTC()
	token.RevokedAt = &now
	_, err := gorm.G[domain.RefreshToken](r.db).
		Where("id = ?", token.ID).
		Updates(context.Background(), *token)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return token, nil
}

func (r *RefreshTokenRepository) DeleteByUserID(userID string) error {
	_, err := gorm.G[domain.RefreshToken](r.db).
		Where("user_id = ?", userID).
		Delete(context.Background())
	if err != nil {
		return util.ParseDatabaseError(err, "idx_refresh_tokens_")
	}
	return nil
}
