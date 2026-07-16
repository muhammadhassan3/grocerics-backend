package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ---------- platform prices ----------

type PlatformPriceRepository struct{ db *gorm.DB }

func NewPlatformPriceRepository(db *gorm.DB) *PlatformPriceRepository {
	return &PlatformPriceRepository{db: db}
}

func (r *PlatformPriceRepository) Upsert(p *domain.PlatformPrice) error {
	err := r.db.WithContext(context.Background()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "variant_id"}, {Name: "platform_id"}, {Name: "city_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"price_paise", "mrp_paise", "available", "inventory", "source", "last_updated_at"}),
		}).Create(p).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_platform_prices_")
	}
	return nil
}

func (r *PlatformPriceRepository) ListByVariantCity(variantID, cityID string) ([]domain.PlatformPrice, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.PlatformPrice](r.db).
		Where("variant_id = ? AND city_id = ?", variantID, cityID).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platform_prices_")
	}
	return items, nil
}

func (r *PlatformPriceRepository) ListByVariantsCity(variantIDs []string, cityID string) ([]domain.PlatformPrice, error) {
	if len(variantIDs) == 0 {
		return nil, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.PlatformPrice](r.db).
		Where("city_id = ? AND variant_id IN ?", cityID, variantIDs).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platform_prices_")
	}
	return items, nil
}

// ---------- variant price summaries (denormalized avg/min) ----------

type VariantPriceSummaryRepository struct{ db *gorm.DB }

func NewVariantPriceSummaryRepository(db *gorm.DB) *VariantPriceSummaryRepository {
	return &VariantPriceSummaryRepository{db: db}
}

func (r *VariantPriceSummaryRepository) Upsert(s *domain.VariantPriceSummary) error {
	err := r.db.WithContext(context.Background()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "variant_id"}, {Name: "city_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"avg_price_paise", "min_price_paise", "min_platform_id", "available_platform_count", "updated_at"}),
		}).Create(s).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_variant_price_summaries_")
	}
	return nil
}

func (r *VariantPriceSummaryRepository) GetMany(variantIDs []string, cityID string) (map[string]domain.VariantPriceSummary, error) {
	out := make(map[string]domain.VariantPriceSummary, len(variantIDs))
	if len(variantIDs) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.VariantPriceSummary](r.db).
		Where("city_id = ? AND variant_id IN ?", cityID, variantIDs).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_variant_price_summaries_")
	}
	for _, s := range items {
		out[s.VariantID] = s
	}
	return out, nil
}
