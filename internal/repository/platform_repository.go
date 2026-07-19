package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
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

func (r *PlatformRepository) ListSearchable() ([]domain.Platform, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Platform](r.db).
		Where("enabled AND deleted_at IS NULL AND qc_name IS NOT NULL AND qc_name <> ''").
		Order("display_order, display_name").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return items, nil
}

func (r *PlatformRepository) ListAdmin(p query.Page, search string) ([]domain.Platform, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Platform](r.db).Where("deleted_at IS NULL")
	if search != "" {
		q = q.Where("(display_name ILIKE ? OR code ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_platforms_")
	}
	items, err := q.Order("display_order, display_name").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return items, total, nil
}

func (r *PlatformRepository) FindByID(id string) (*domain.Platform, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Platform](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return &data, nil
}

func (r *PlatformRepository) Create(p *domain.Platform) (*domain.Platform, error) {
	if err := gorm.G[domain.Platform](r.db).Create(context.Background(), p); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return p, nil
}

func (r *PlatformRepository) Update(id string, fields map[string]any) (*domain.Platform, error) {
	if err := adminUpdateFields[domain.Platform](r.db, id, fields, "idx_platforms_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *PlatformRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Platform](r.db, id, adminID, "idx_platforms_")
}

func (r *PlatformRepository) Reorder(ids []string) error {
	return adminReorder[domain.Platform](r.db, ids, "idx_platforms_")
}

func (r *PlatformRepository) FindByIDs(ids []string) (map[string]domain.Platform, error) {
	out := make(map[string]domain.Platform, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.Platform](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_platforms_")
	}
	for _, p := range items {
		out[p.ID] = p
	}
	return out, nil
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
	if l.ImageURL != nil { // keep the last known image if this link carries none
		existing.ImageURL = l.ImageURL
	}
	if _, err := gorm.G[domain.ProductPlatformLink](r.db).Where("id = ?", existing.ID).Updates(ctx, *existing); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	return existing, nil
}

func (r *ProductPlatformLinkRepository) PrimaryImagesByVariants(variantIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(variantIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		VariantID string
		ImageURL  string
	}
	err := r.db.WithContext(context.Background()).
		Raw(`SELECT DISTINCT ON (l.variant_id) l.variant_id, l.image_url
		     FROM product_platform_links l
		     JOIN platforms p ON p.id = l.platform_id AND p.deleted_at IS NULL
		     WHERE l.variant_id IN ? AND l.deleted_at IS NULL
		       AND l.image_url IS NOT NULL AND l.image_url <> ''
		     ORDER BY l.variant_id, p.display_order ASC`, variantIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_platform_links_")
	}
	for _, row := range rows {
		out[row.VariantID] = row.ImageURL
	}
	return out, nil
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
