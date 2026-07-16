package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

func (r *ProductRepository) Create(p *domain.Product) (*domain.Product, error) {
	if err := gorm.G[domain.Product](r.db).Create(context.Background(), p); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_products_")
	}
	return p, nil
}

func (r *ProductRepository) Update(id string, fields map[string]any) (*domain.Product, error) {
	if err := adminUpdateFields[domain.Product](r.db, id, fields, "idx_products_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *ProductRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Product](r.db, id, adminID, "idx_products_")
}

func (r *ProductRepository) ListAdmin(p query.Page, search string) ([]domain.Product, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Product](r.db).Where("deleted_at IS NULL")
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_products_")
	}
	items, err := q.Order("created_at DESC, id DESC").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_products_")
	}
	return items, total, nil
}

func (r *ProductRepository) CountVariants(ids []string) (map[string]int, error) {
	return countBy(r.db, "product_variants", "product_id", ids, "idx_product_variants_")
}

type InventoryStats struct {
	Categories int64
	Products   int64
	Brands     int64
	Platforms  int64
}

func (r *ProductRepository) Stats() (InventoryStats, error) {
	ctx := context.Background()
	var s InventoryStats
	count := func(table, extra string) (int64, error) {
		var n int64
		q := r.db.WithContext(ctx).Table(table).Where("deleted_at IS NULL")
		if extra != "" {
			q = q.Where(extra)
		}
		return n, q.Count(&n).Error
	}
	var err error
	if s.Categories, err = count("categories", ""); err != nil {
		return s, util.ParseDatabaseError(err, "idx_categories_")
	}
	if s.Products, err = count("products", ""); err != nil {
		return s, util.ParseDatabaseError(err, "idx_products_")
	}
	if s.Brands, err = count("brands", ""); err != nil {
		return s, util.ParseDatabaseError(err, "idx_brands_")
	}
	if s.Platforms, err = count("platforms", "enabled"); err != nil {
		return s, util.ParseDatabaseError(err, "idx_platforms_")
	}
	return s, nil
}

func (r *CategoryRepository) NamesByIDs(ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	items, err := gorm.G[domain.Category](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(context.Background())
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_categories_")
	}
	for _, c := range items {
		out[c.ID] = c.Name
	}
	return out, nil
}

func (r *ProductVariantRepository) Create(v *domain.ProductVariant) (*domain.ProductVariant, error) {
	if err := gorm.G[domain.ProductVariant](r.db).Create(context.Background(), v); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return v, nil
}

func (r *ProductVariantRepository) Update(id string, fields map[string]any) (*domain.ProductVariant, error) {
	if err := adminUpdateFields[domain.ProductVariant](r.db, id, fields, "idx_product_variants_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *ProductVariantRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.ProductVariant](r.db, id, adminID, "idx_product_variants_")
}
