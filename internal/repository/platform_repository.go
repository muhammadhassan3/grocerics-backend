package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ---------- platforms ----------

type PlatformRepository struct{ db *gorm.DB }

func NewPlatformRepository(db *gorm.DB) *PlatformRepository { return &PlatformRepository{db: db} }

func (r *PlatformRepository) ListEnabled() ([]domain.Platform, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Platform](r.db).
		Where("enabled AND deleted_at IS NULL").
		Order("display_order, display_name").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return items, nil
}

func (r *PlatformRepository) FindByCode(code string) (*domain.Platform, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Platform](r.db).Where("code = ? AND deleted_at IS NULL", code).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return &data, nil
}

// ---------- product ↔ platform links ----------

type ProductPlatformLinkRepository struct{ db *gorm.DB }

func NewProductPlatformLinkRepository(db *gorm.DB) *ProductPlatformLinkRepository {
	return &ProductPlatformLinkRepository{db: db}
}

func (r *ProductPlatformLinkRepository) FindByVariantPlatform(variantID, platformID string) (*domain.ProductPlatformLink, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.ProductPlatformLink](r.db).
		Where("variant_id = ? AND platform_id = ? AND deleted_at IS NULL", variantID, platformID).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	return &data, nil
}

func (r *ProductPlatformLinkRepository) Upsert(l *domain.ProductPlatformLink) (*domain.ProductPlatformLink, error) {
	existing, err := r.FindByVariantPlatform(l.VariantID, l.PlatformID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if existing == nil {
		if err := gorm.G[domain.ProductPlatformLink](r.db).Create(ctx, l); err != nil {
			return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
		}
		return l, nil
	}
	existing.PlatformSKU = l.PlatformSKU
	existing.ProductURL = l.ProductURL
	existing.DeepLink = l.DeepLink
	if _, err := gorm.G[domain.ProductPlatformLink](r.db).Where("id = ?", existing.ID).Updates(ctx, *existing); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	return existing, nil
}

func (r *ProductPlatformLinkRepository) ListByVariant(variantID string) ([]domain.ProductPlatformLink, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.ProductPlatformLink](r.db).
		Where("variant_id = ? AND deleted_at IS NULL", variantID).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	return items, nil
}

func (r *ProductPlatformLinkRepository) ListAll() ([]domain.ProductPlatformLink, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.ProductPlatformLink](r.db).Where("deleted_at IS NULL").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	return items, nil
}

// ---------- delivery ETAs (pincode-level) ----------

type PlatformDeliveryETARepository struct{ db *gorm.DB }

func NewPlatformDeliveryETARepository(db *gorm.DB) *PlatformDeliveryETARepository {
	return &PlatformDeliveryETARepository{db: db}
}

func (r *PlatformDeliveryETARepository) Upsert(e *domain.PlatformDeliveryETA) error {
	err := r.db.WithContext(context.Background()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "platform_id"}, {Name: "pincode"}},
			DoUpdates: clause.AssignmentColumns([]string{"eta_minutes", "serviceable", "last_updated_at"}),
		}).Create(e).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_platform_delivery_etas_")
	}
	return nil
}

func (r *PlatformDeliveryETARepository) ListByPincode(pincode string) ([]domain.PlatformDeliveryETA, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.PlatformDeliveryETA](r.db).Where("pincode = ?", pincode).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platform_delivery_etas_")
	}
	return items, nil
}
