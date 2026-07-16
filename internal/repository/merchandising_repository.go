package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

// ---------- banners ----------

type BannerRepository struct{ db *gorm.DB }

func NewBannerRepository(db *gorm.DB) *BannerRepository { return &BannerRepository{db: db} }

func (r *BannerRepository) ListActive() ([]domain.Banner, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Banner](r.db).
		Where("is_active AND deleted_at IS NULL").
		Where("(start_date IS NULL OR start_date <= now())").
		Where("(end_date IS NULL OR end_date >= now())").
		Order("created_at DESC").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_banners_")
	}
	return items, nil
}
