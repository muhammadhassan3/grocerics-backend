package config

import (
	"errors"
	"fmt"

	"grocerics-backend/internal/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func sptr(v string) *string   { return &v }
func fptr(v float64) *float64 { return &v }

// ensure inserts row unless a live one already matches the natural key, and
// returns whichever exists now. Insert-only by design: admins reorder and disable
// things through the admin UI, and re-running the seeder must never undo that.
func ensure[T any](db *gorm.DB, row *T, where string, args ...any) (found *T, created bool, err error) {
	var existing T
	err = db.Where(where+" AND deleted_at IS NULL", args...).First(&existing).Error
	if err == nil {
		return &existing, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if err := db.Create(row).Error; err != nil {
		return nil, false, err
	}
	return row, true, nil
}

// SeedReference seeds configuration data only: cities, platforms, categories,
// subcategories, brands. It deliberately creates no products, variants, links or
// prices -- the catalog is built by admins through the linker against real
// QuickCommerce data, not invented here.
//
// Runs only via `api -seed`, so it carries no ENV guard: invoking it IS the
// intent. That is what lets it configure a production box, which the old
// env-gated demo seeder could never do.
func SeedReference(db *gorm.DB) error {
	created := 0
	note := func(kind, name string, made bool) {
		if made {
			created++
			zap.S().Infow("seed: created", "kind", kind, "name", name)
		}
	}

	// --- cities ---
	// lat/lng/default_pincode are the QuickCommerce search anchor for the city.
	// Each was verified live to return real results on BlinkIt/Zepto/Swiggy; a city
	// without them is unusable by the linker (qcLocation returns CITY_NO_LOCATION).
	// enabled=false would mean "seeding lens only": usable to harvest item ids, but
	// hidden from users and skipped by the price fan-out.
	cityDefs := []struct {
		name, slug, state, pincode string
		lat, lng                   float64
	}{
		{"Delhi", "delhi", "Delhi", "110035", 28.6980, 77.1490},
		{"Mumbai", "mumbai", "Maharashtra", "400050", 19.0596, 72.8295},
		{"Bangalore", "bangalore", "Karnataka", "560034", 12.9352, 77.6245},
	}
	for _, d := range cityDefs {
		_, made, err := ensure(db, &domain.City{
			Name: d.name, Slug: d.slug, State: sptr(d.state),
			Lat: fptr(d.lat), Lng: fptr(d.lng), DefaultPincode: sptr(d.pincode),
			Enabled: true,
		}, "slug = ?", d.slug)
		if err != nil {
			return fmt.Errorf("city %s: %w", d.slug, err)
		}
		note("city", d.name, made)
	}

	// --- platforms ---
	// qc is the QuickCommerce platform name; blank makes a platform unsearchable.
	// All six verified against GET /supported-platforms.
	platformDefs := []struct {
		code, name, qc, eta string
		enabled             bool
		order               int
	}{
		{"zepto", "Zepto", "Zepto", "12 Mins", true, 0},
		{"blinkit", "Blinkit", "BlinkIt", "10 Mins", true, 1},
		{"instamart", "Swiggy Instamart", "Swiggy", "15 Mins", true, 2},
		{"flipkart", "Flipkart Minutes", "Minutes", "14 Mins", false, 3},
		{"jiomart", "JioMart", "JioMart", "20 Mins", false, 4},
		{"amazon_now", "Amazon Now", "Amazon", "18 Mins", false, 5},
	}
	for _, d := range platformDefs {
		_, made, err := ensure(db, &domain.Platform{
			Code: d.code, DisplayName: d.name, QCName: sptr(d.qc),
			DeliveryETAText: sptr(d.eta), Enabled: d.enabled, DisplayOrder: d.order,
		}, "code = ?", d.code)
		if err != nil {
			return fmt.Errorf("platform %s: %w", d.code, err)
		}
		note("platform", d.code, made)
	}

	// --- categories ---
	catDefs := []struct {
		name, slug string
		top        bool
		order      int
	}{
		{"Groceries", "groceries", true, 0},
		{"Snacks", "snacks", true, 1},
		{"Dairy", "dairy", false, 2},
		{"Beverages", "beverages", true, 3},
	}
	cats := make(map[string]*domain.Category, len(catDefs))
	for _, d := range catDefs {
		row, made, err := ensure(db, &domain.Category{
			Name: d.name, Slug: d.slug, IsTopCategory: d.top,
			Status: domain.StatusActive, DisplayOrder: d.order,
		}, "slug = ?", d.slug)
		if err != nil {
			return fmt.Errorf("category %s: %w", d.slug, err)
		}
		cats[d.slug] = row
		note("category", d.name, made)
	}

	// --- subcategories ---
	// No unique index exists beyond the PK, so the natural key (category_id, name)
	// is enforced here rather than by the database.
	subDefs := []struct {
		name, slug, cat string
		top             bool
		order           int
	}{
		{"Soft Drink", "soft-drink", "beverages", true, 0},
	}
	for _, d := range subDefs {
		parent, ok := cats[d.cat]
		if !ok {
			return fmt.Errorf("subcategory %s: parent category %q missing", d.name, d.cat)
		}
		_, made, err := ensure(db, &domain.Subcategory{
			CategoryID: parent.ID, Name: d.name, Slug: sptr(d.slug),
			IsTopSubcategory: d.top, Status: domain.StatusActive, DisplayOrder: d.order,
		}, "category_id = ? AND name = ?", parent.ID, d.name)
		if err != nil {
			return fmt.Errorf("subcategory %s: %w", d.name, err)
		}
		note("subcategory", d.name, made)
	}

	// --- brands ---
	// No unique index here either, so name is the natural key.
	brandDefs := []struct {
		name, slug string
		top        bool
		order      int
	}{
		{"Amul", "amul", true, 0},
		{"Britannia", "britannia", true, 1},
		{"Sunfeast", "sunfeast", true, 2},
		{"Coca-Cola", "coca-cola", true, 3},
		{"Thums Up", "thums-up", true, 4},
	}
	for _, d := range brandDefs {
		_, made, err := ensure(db, &domain.Brand{
			Name: d.name, Slug: sptr(d.slug), IsTopBrand: d.top,
			Status: domain.StatusActive, DisplayOrder: d.order,
		}, "name = ?", d.name)
		if err != nil {
			return fmt.Errorf("brand %s: %w", d.name, err)
		}
		note("brand", d.name, made)
	}

	zap.S().Infow("seed: reference data ready",
		"created", created,
		"unchanged", len(cityDefs)+len(platformDefs)+len(catDefs)+len(subDefs)+len(brandDefs)-created)
	return nil
}
