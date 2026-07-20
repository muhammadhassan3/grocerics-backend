package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

// ---------- categories ----------

type CategoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *CategoryRepository) ListVisible(topOnly bool) ([]domain.Category, error) {
	ctx := context.Background()
	q := gorm.G[domain.Category](r.db).Where("status = 'active' AND deleted_at IS NULL")
	if topOnly {
		q = q.Where("is_top_category")
	}
	items, err := q.Order("display_order, name").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_categories_")
	}
	return items, nil
}

func (r *CategoryRepository) FindByID(id string) (*domain.Category, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Category](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_categories_")
	}
	return &data, nil
}

// ---------- subcategories ----------

type SubcategoryRepository struct{ db *gorm.DB }

func NewSubcategoryRepository(db *gorm.DB) *SubcategoryRepository {
	return &SubcategoryRepository{db: db}
}

func (r *SubcategoryRepository) ListByCategory(categoryID string) ([]domain.Subcategory, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Subcategory](r.db).
		Where("category_id = ? AND status = 'active' AND deleted_at IS NULL", categoryID).
		Order("display_order, name").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	return items, nil
}

// ---------- brands ----------

type BrandRepository struct{ db *gorm.DB }

func NewBrandRepository(db *gorm.DB) *BrandRepository { return &BrandRepository{db: db} }

func (r *BrandRepository) FindByIDs(ids []string) (map[string]domain.Brand, error) {
	out := make(map[string]domain.Brand, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.Brand](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_brands_")
	}
	for _, b := range items {
		out[b.ID] = b
	}
	return out, nil
}

// ---------- products ----------

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }

func (r *ProductRepository) FindByID(id string) (*domain.Product, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Product](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_products_")
	}
	return &data, nil
}

func (r *ProductRepository) FindByIDs(ids []string) (map[string]domain.Product, error) {
	out := make(map[string]domain.Product, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.Product](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_products_")
	}
	for _, p := range items {
		out[p.ID] = p
	}
	return out, nil
}

func (r *ProductRepository) ListByCategory(categoryID string, p query.Page) ([]domain.Product, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Product](r.db).Where("category_id = ? AND status = 'active' AND deleted_at IS NULL", categoryID)
	return paginateProducts(ctx, q, p)
}

// SearchByNameOrBrand matches active products whose name OR brand name contains
// the term (case-insensitive, partial). Powers the variant search screen.
func (r *ProductRepository) SearchByNameOrBrand(term string, p query.Page) ([]domain.Product, int64, error) {
	ctx := context.Background()
	like := "%" + term + "%"
	var brandIDs []string
	_ = r.db.WithContext(ctx).Model(&domain.Brand{}).
		Where("name ILIKE ? AND deleted_at IS NULL", like).Pluck("id", &brandIDs)

	q := gorm.G[domain.Product](r.db).Where("status = 'active' AND deleted_at IS NULL")
	if len(brandIDs) > 0 {
		q = q.Where("name ILIKE ? OR brand_id IN ?", like, brandIDs)
	} else {
		q = q.Where("name ILIKE ?", like)
	}
	return paginateProducts(ctx, q, p)
}

func (r *ProductRepository) ListTop(limit int) ([]domain.Product, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Product](r.db).
		Where("is_top_item AND status = 'active' AND deleted_at IS NULL").
		Order("created_at DESC").Limit(limit).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_products_")
	}
	return items, nil
}

func (r *ProductRepository) ListSimilar(categoryID, excludeProductID string, limit int) ([]domain.Product, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Product](r.db).
		Where("category_id = ? AND id <> ? AND status = 'active' AND deleted_at IS NULL", categoryID, excludeProductID).
		Order("created_at DESC").Limit(limit).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_products_")
	}
	return items, nil
}

func (r *ProductRepository) ListDealVariantIDs(cityID string, limit int) ([]string, error) {
	ctx := context.Background()
	var ids []string
	err := r.db.WithContext(ctx).
		Table("platform_prices pp").
		Joins("JOIN product_variants v ON v.id = pp.variant_id").
		Joins("JOIN products p ON p.id = v.product_id").
		Where("pp.city_id = ? AND pp.available AND pp.mrp_paise IS NOT NULL AND pp.mrp_paise > pp.price_paise", cityID).
		Where("p.status = 'active' AND p.deleted_at IS NULL AND v.deleted_at IS NULL").
		Distinct("v.id").Limit(limit).Pluck("v.id", &ids).Error
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return ids, nil
}

func paginateProducts(ctx context.Context, q gorm.ChainInterface[domain.Product], p query.Page) ([]domain.Product, int64, error) {
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

// ---------- product variants ----------

type ProductVariantRepository struct{ db *gorm.DB }

func NewProductVariantRepository(db *gorm.DB) *ProductVariantRepository {
	return &ProductVariantRepository{db: db}
}

func (r *ProductVariantRepository) ListByProducts(productIDs []string) ([]domain.ProductVariant, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.ProductVariant](r.db).
		Where("product_id IN ? AND deleted_at IS NULL", productIDs).
		Order("product_id, display_order, volume_value").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return items, nil
}

func (r *ProductVariantRepository) ListByProduct(productID string) ([]domain.ProductVariant, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.ProductVariant](r.db).
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Order("display_order, volume_value").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return items, nil
}

func (r *ProductVariantRepository) FindByID(id string) (*domain.ProductVariant, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.ProductVariant](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return &data, nil
}

func (r *ProductVariantRepository) DefaultsForProducts(productIDs []string) (map[string]domain.ProductVariant, error) {
	out := make(map[string]domain.ProductVariant, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.ProductVariant](r.db).
		Where("product_id IN ? AND deleted_at IS NULL", productIDs).
		Order("display_order, volume_value").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	for _, v := range items {
		if _, seen := out[v.ProductID]; !seen {
			out[v.ProductID] = v
		}
	}
	return out, nil
}

// returns the requested variants queriedd by id
func (r *ProductVariantRepository) FindByIDs(ids []string) (map[string]domain.ProductVariant, error) {
	out := make(map[string]domain.ProductVariant, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ctx := context.Background()
	items, err := gorm.G[domain.ProductVariant](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	for _, v := range items {
		out[v.ID] = v
	}
	return out, nil
}

func (r *ProductVariantRepository) ListByIDsOrdered(ids []string) ([]domain.ProductVariant, error) {
	m, err := r.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProductVariant, 0, len(ids))
	for _, id := range ids {
		if v, ok := m[id]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}

func (r *ProductVariantRepository) ListAll() ([]domain.ProductVariant, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.ProductVariant](r.db).Where("deleted_at IS NULL").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_variants_")
	}
	return items, nil
}

// ---------- product images ----------

type ProductImageRepository struct{ db *gorm.DB }

func NewProductImageRepository(db *gorm.DB) *ProductImageRepository {
	return &ProductImageRepository{db: db}
}

func (r *ProductImageRepository) ListByProduct(productID string) ([]domain.ProductImage, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.ProductImage](r.db).
		Where("product_id = ? AND deleted_at IS NULL", productID).
		Order("display_order").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_product_images_")
	}
	return items, nil
}
