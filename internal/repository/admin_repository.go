package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

func adminUpdateFields[T any](db *gorm.DB, id string, fields map[string]any, idx string) error {
	if len(fields) == 0 {
		return nil
	}
	var zero T
	err := db.WithContext(context.Background()).
		Model(&zero).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error
	return util.ParseDatabaseError(err, idx)
}

func adminReorder[T any](db *gorm.DB, ids []string, idx string) error {
	if len(ids) == 0 {
		return nil
	}
	err := db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var zero T
		for i, id := range ids {
			if err := tx.Model(&zero).Where("id = ? AND deleted_at IS NULL", id).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return util.ParseDatabaseError(err, idx)
}

func adminSoftDelete[T any](db *gorm.DB, id, adminID, idx string) error {
	fields := map[string]any{"deleted_at": gorm.Expr("now()")}
	if adminID != "" {
		fields["deleted_by"] = adminID
	}
	var zero T
	err := db.WithContext(context.Background()).
		Model(&zero).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error
	return util.ParseDatabaseError(err, idx)
}

func countBy(db *gorm.DB, table, col string, ids []string, idx string) (map[string]int, error) {
	out := make(map[string]int, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		Key string
		Cnt int
	}
	err := db.WithContext(context.Background()).
		Table(table).
		Select(col+" AS key, count(*) AS cnt").
		Where(col+" IN ? AND deleted_at IS NULL", ids).
		Group(col).Scan(&rows).Error
	if err != nil {
		return nil, util.ParseDatabaseError(err, idx)
	}
	for _, r := range rows {
		out[r.Key] = r.Cnt
	}
	return out, nil
}

func (r *CategoryRepository) Create(c *domain.Category) (*domain.Category, error) {
	if err := gorm.G[domain.Category](r.db).Create(context.Background(), c); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_categories_")
	}
	return c, nil
}

func (r *CategoryRepository) Update(id string, fields map[string]any) (*domain.Category, error) {
	if err := adminUpdateFields[domain.Category](r.db, id, fields, "idx_categories_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *CategoryRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Category](r.db, id, adminID, "idx_categories_")
}

func (r *CategoryRepository) Reorder(ids []string) error {
	return adminReorder[domain.Category](r.db, ids, "idx_categories_")
}

func (r *CategoryRepository) ListAdmin(p query.Page, search string, hasProducts bool) ([]domain.Category, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Category](r.db).Where("deleted_at IS NULL")
	if hasProducts {
		q = q.Where(`EXISTS (SELECT 1 FROM products p JOIN product_variants v ON v.product_id = p.id
			WHERE p.category_id = categories.id AND p.status = 'active' AND p.deleted_at IS NULL AND v.deleted_at IS NULL)`)
	}
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_categories_")
	}
	items, err := q.Order("display_order, name").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_categories_")
	}
	return items, total, nil
}

func (r *CategoryRepository) CountSubcategories(ids []string) (map[string]int, error) {
	return countBy(r.db, "subcategories", "category_id", ids, "idx_subcategories_")
}

func (r *SubcategoryRepository) Create(s *domain.Subcategory) (*domain.Subcategory, error) {
	if err := gorm.G[domain.Subcategory](r.db).Create(context.Background(), s); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	return s, nil
}

func (r *SubcategoryRepository) Update(id string, fields map[string]any) (*domain.Subcategory, error) {
	if err := adminUpdateFields[domain.Subcategory](r.db, id, fields, "idx_subcategories_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *SubcategoryRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Subcategory](r.db, id, adminID, "idx_subcategories_")
}

func (r *SubcategoryRepository) Reorder(ids []string) error {
	return adminReorder[domain.Subcategory](r.db, ids, "idx_subcategories_")
}

func (r *SubcategoryRepository) FindByID(id string) (*domain.Subcategory, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Subcategory](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	return &data, nil
}
func (r *SubcategoryRepository) NamesByIDs(ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	items, err := gorm.G[domain.Subcategory](r.db).Where("id IN ? AND deleted_at IS NULL", ids).Find(context.Background())
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	for _, s := range items {
		out[s.ID] = s.Name
	}
	return out, nil
}

func (r *SubcategoryRepository) ListAdmin(p query.Page, categoryID, search string, hasProducts bool) ([]domain.Subcategory, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Subcategory](r.db).Where("deleted_at IS NULL")
	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}
	if hasProducts {
		q = q.Where(`EXISTS (SELECT 1 FROM products p JOIN product_variants v ON v.product_id = p.id
			WHERE p.subcategory_id = subcategories.id AND p.status = 'active' AND p.deleted_at IS NULL AND v.deleted_at IS NULL)`)
	}
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	items, err := q.Order("display_order, name").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_subcategories_")
	}
	return items, total, nil
}

func (r *BrandRepository) Create(b *domain.Brand) (*domain.Brand, error) {
	if err := gorm.G[domain.Brand](r.db).Create(context.Background(), b); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_brands_")
	}
	return b, nil
}

func (r *BrandRepository) Update(id string, fields map[string]any) (*domain.Brand, error) {
	if err := adminUpdateFields[domain.Brand](r.db, id, fields, "idx_brands_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *BrandRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Brand](r.db, id, adminID, "idx_brands_")
}

func (r *BrandRepository) Reorder(ids []string) error {
	return adminReorder[domain.Brand](r.db, ids, "idx_brands_")
}

func (r *BrandRepository) FindByID(id string) (*domain.Brand, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Brand](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_brands_")
	}
	return &data, nil
}

func (r *BrandRepository) ListAdmin(p query.Page, search string) ([]domain.Brand, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Brand](r.db).Where("deleted_at IS NULL")
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_brands_")
	}
	items, err := q.Order("display_order, name").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_brands_")
	}
	return items, total, nil
}

func (r *BrandRepository) CountProducts(ids []string) (map[string]int, error) {
	return countBy(r.db, "products", "brand_id", ids, "idx_products_")
}

func (r *BannerRepository) Create(b *domain.Banner) (*domain.Banner, error) {
	if err := gorm.G[domain.Banner](r.db).Create(context.Background(), b); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_banners_")
	}
	return b, nil
}

func (r *BannerRepository) Update(id string, fields map[string]any) (*domain.Banner, error) {
	if err := adminUpdateFields[domain.Banner](r.db, id, fields, "idx_banners_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *BannerRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.Banner](r.db, id, adminID, "idx_banners_")
}

func (r *BannerRepository) FindByID(id string) (*domain.Banner, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Banner](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_banners_")
	}
	return &data, nil
}

func (r *BannerRepository) ListAdmin(p query.Page) ([]domain.Banner, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.Banner](r.db).Where("deleted_at IS NULL")
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_banners_")
	}
	items, err := q.Order("created_at DESC").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_banners_")
	}
	return items, total, nil
}

func (r *CityRepository) Create(c *domain.City) (*domain.City, error) {
	if err := gorm.G[domain.City](r.db).Create(context.Background(), c); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_cities_")
	}
	return c, nil
}

func (r *CityRepository) Update(id string, fields map[string]any) (*domain.City, error) {
	if err := adminUpdateFields[domain.City](r.db, id, fields, "idx_cities_"); err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *CityRepository) SoftDelete(id, adminID string) error {
	return adminSoftDelete[domain.City](r.db, id, adminID, "idx_cities_")
}

func (r *CityRepository) ListAdmin(p query.Page, search string) ([]domain.City, int64, error) {
	ctx := context.Background()
	q := gorm.G[domain.City](r.db).Where("deleted_at IS NULL")
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_cities_")
	}
	items, err := q.Order("name").Limit(p.Limit()).Offset(p.Offset()).Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_cities_")
	}
	return items, total, nil
}
