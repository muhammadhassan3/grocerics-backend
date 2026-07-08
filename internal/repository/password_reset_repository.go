package repository

import (
	"context"
	"time"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(reset *domain.PasswordReset) (*domain.PasswordReset, error) {
	err := gorm.G[domain.PasswordReset](r.db).Create(context.Background(), reset)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_password_resets_")
	}
	return reset, nil
}

// FindByTokenHash returns (nil, nil) when no row matches, matching
// UserRepository.FindByID's convention.
func (r *PasswordResetRepository) FindByTokenHash(hash string) (*domain.PasswordReset, error) {
	data, err := gorm.G[domain.PasswordReset](r.db).
		Where("token_hash = ?", hash).
		First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_password_resets_")
	}
	return &data, nil
}

func (r *PasswordResetRepository) MarkUsed(reset *domain.PasswordReset) error {
	now := time.Now().UTC()
	reset.UsedAt = &now
	_, err := gorm.G[domain.PasswordReset](r.db).
		Where("id = ?", reset.ID).
		Updates(context.Background(), *reset)
	if err != nil {
		return util.ParseDatabaseError(err, "idx_password_resets_")
	}
	return nil
}
