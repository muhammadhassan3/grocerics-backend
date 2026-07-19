package repository

import (
	"context"
	"time"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type AdminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) Create(a *domain.Admin) (*domain.Admin, error) {
	if err := gorm.G[domain.Admin](r.db).Create(context.Background(), a); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_admins_")
	}
	return a, nil
}

func (r *AdminRepository) FindByEmail(email string) (*domain.Admin, error) {
	data, err := gorm.G[domain.Admin](r.db).
		Where("email = ? AND deleted_at IS NULL", email).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_admins_")
	}
	return &data, nil
}

func (r *AdminRepository) FindByID(id string) (*domain.Admin, error) {
	data, err := gorm.G[domain.Admin](r.db).
		Where("id = ? AND deleted_at IS NULL", id).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_admins_")
	}
	return &data, nil
}

func (r *AdminRepository) Update(a *domain.Admin) (*domain.Admin, error) {
	if _, err := gorm.G[domain.Admin](r.db).Where("id = ?", a.ID).Updates(context.Background(), *a); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_admins_")
	}
	return a, nil
}

type AdminRefreshTokenRepository struct{ db *gorm.DB }

func NewAdminRefreshTokenRepository(db *gorm.DB) *AdminRefreshTokenRepository {
	return &AdminRefreshTokenRepository{db: db}
}

func (r *AdminRefreshTokenRepository) Create(t *domain.AdminRefreshToken) (*domain.AdminRefreshToken, error) {
	if err := gorm.G[domain.AdminRefreshToken](r.db).Create(context.Background(), t); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_admin_refresh_tokens_")
	}
	return t, nil
}

func (r *AdminRefreshTokenRepository) FindByTokenHash(hash string) (*domain.AdminRefreshToken, error) {
	data, err := gorm.G[domain.AdminRefreshToken](r.db).
		Where("token_hash = ?", hash).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_admin_refresh_tokens_")
	}
	return &data, nil
}

func (r *AdminRefreshTokenRepository) FindByTokenHashAndAdminID(hash, adminID string) (*domain.AdminRefreshToken, error) {
	data, err := gorm.G[domain.AdminRefreshToken](r.db).
		Where("token_hash = ? AND admin_id = ?", hash, adminID).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_admin_refresh_tokens_")
	}
	return &data, nil
}

func (r *AdminRefreshTokenRepository) Update(t *domain.AdminRefreshToken) (*domain.AdminRefreshToken, error) {
	if _, err := gorm.G[domain.AdminRefreshToken](r.db).Where("id = ?", t.ID).Updates(context.Background(), *t); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_admin_refresh_tokens_")
	}
	return t, nil
}

func (r *AdminRefreshTokenRepository) Revoke(t *domain.AdminRefreshToken) error {
	now := time.Now().UTC()
	t.RevokedAt = &now
	if _, err := gorm.G[domain.AdminRefreshToken](r.db).Where("id = ?", t.ID).Updates(context.Background(), *t); err != nil {
		return util.ParseDatabaseError(err, "idx_admin_refresh_tokens_")
	}
	return nil
}

type AdminPasswordResetRepository struct{ db *gorm.DB }

func NewAdminPasswordResetRepository(db *gorm.DB) *AdminPasswordResetRepository {
	return &AdminPasswordResetRepository{db: db}
}

func (r *AdminPasswordResetRepository) Create(p *domain.AdminPasswordReset) (*domain.AdminPasswordReset, error) {
	if err := gorm.G[domain.AdminPasswordReset](r.db).Create(context.Background(), p); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_admin_password_resets_")
	}
	return p, nil
}

func (r *AdminPasswordResetRepository) FindByTokenHash(hash string) (*domain.AdminPasswordReset, error) {
	data, err := gorm.G[domain.AdminPasswordReset](r.db).
		Where("token_hash = ?", hash).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_admin_password_resets_")
	}
	return &data, nil
}

func (r *AdminPasswordResetRepository) MarkUsed(p *domain.AdminPasswordReset) error {
	now := time.Now().UTC()
	p.UsedAt = &now
	if _, err := gorm.G[domain.AdminPasswordReset](r.db).Where("id = ?", p.ID).Updates(context.Background(), *p); err != nil {
		return util.ParseDatabaseError(err, "idx_admin_password_resets_")
	}
	return nil
}
