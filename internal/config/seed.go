package config

import (
	"fmt"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func sptr(v string) *string   { return &v }
func fptr(v float64) *float64 { return &v }
func i64ptr(v int64) *int64   { return &v }

var seedPlatformOffset = map[string]int64{
	"blinkit": 0, "zepto": 500, "instamart": 1000,
	"flipkart": -500, "jiomart": 1500, "amazon_now": 2000,
}

// SeedDemo populates reference data (city, platforms, a small catalog, links,
// prices, ETAs) so the consumer endpoints render end-to-end. Dev/staging only,
// and idempotent — it no-ops once a city exists.
func SeedDemo(db *gorm.DB, env string) {
	if env == "production" {
		return
	}
	var cityCount int64
	if err := db.Model(&domain.City{}).Count(&cityCount).Error; err != nil {
		zap.S().Warnw("seed: city count failed", "error", err)
		return
	}
	if cityCount > 0 {
		return // already seeded
	}

	// --- geography ---
	// Delhi's QC anchor: the serviceable pincode 110035 and its precise lat/lng
	// (28.6980/77.1490 responds on all platforms incl. Zepto, which needs a pincode).
	pincode := "110035"
	city := &domain.City{Name: "Delhi", Slug: "delhi", Lat: fptr(28.6980), Lng: fptr(77.1490), DefaultPincode: sptr(pincode), Enabled: true, DisplayOrder: 1}
	if err := db.Create(city).Error; err != nil {
		zap.S().Warnw("seed: city failed", "error", err)
		return
	}
	db.Create(&domain.Pincode{Pincode: pincode, CityID: city.ID, Lat: fptr(28.6980), Lng: fptr(77.1490), Serviceable: true})

	// --- platforms ---
	// qc is the QuickCommerce platform name (empty = not tracked via QC).
	platformDefs := []struct{ code, name, eta, qc string }{
		{"blinkit", "Blinkit", "10 Mins", "BlinkIt"},
		{"zepto", "Zepto", "12 Mins", "Zepto"},
		{"instamart", "Swiggy Instamart", "15 Mins", "Swiggy"},
		{"flipkart", "Flipkart Minutes", "14 Mins", "Minutes"},
		{"jiomart", "JioMart", "20 Mins", "JioMart"},
		{"amazon_now", "Amazon Now", "18 Mins", "Amazon"},
	}
	platforms := make([]domain.Platform, 0, len(platformDefs))
	for i, d := range platformDefs {
		p := domain.Platform{Code: d.code, DisplayName: d.name, QCName: sptr(d.qc), DeliveryETAText: sptr(d.eta), Enabled: true, DisplayOrder: i + 1}
		if err := db.Create(&p).Error; err != nil {
			zap.S().Warnw("seed: platform failed", "code", d.code, "error", err)
			continue
		}
		platforms = append(platforms, p)
		// pincode-level ETA
		eta := 10 + i*2
		db.Create(&domain.PlatformDeliveryETA{PlatformID: p.ID, Pincode: pincode, ETAMinutes: &eta, Serviceable: true})
	}

	// --- catalog ---
	catDefs := []struct {
		name, slug string
		top        bool
	}{
		{"Snacks", "snacks", true}, {"Beverages", "beverages", true},
		{"Groceries", "groceries", true}, {"Dairy", "dairy", false},
	}
	cats := map[string]domain.Category{}
	for i, d := range catDefs {
		c := domain.Category{Name: d.name, Slug: d.slug, IsTopCategory: d.top, Status: domain.StatusActive, DisplayOrder: i + 1}
		db.Create(&c)
		cats[d.slug] = c
	}
	brandDefs := []string{"Sunfeast", "Amul", "Britannia"}
	brands := map[string]domain.Brand{}
	for _, name := range brandDefs {
		b := domain.Brand{Name: name, IsTopBrand: true, Status: domain.StatusActive}
		db.Create(&b)
		brands[name] = b
	}

	// products with variants and a base price (paise) per variant
	type variantDef struct {
		vol   float64
		unit  domain.VolumeUnit
		price int64
	}
	prodDefs := []struct {
		name, cat, brand string
		top              bool
		variants         []variantDef
	}{
		{"Sunfeast Whole Grain Bread", "snacks", "Sunfeast", true, []variantDef{{250, domain.VolumeUnitGm, 25000}, {500, domain.VolumeUnitGm, 30000}}},
		{"Amul Taaza Toned Milk", "dairy", "Amul", true, []variantDef{{500, domain.VolumeUnitMl, 3300}, {1000, domain.VolumeUnitMl, 6600}}},
		{"Britannia Bourbon Biscuits", "snacks", "Britannia", false, []variantDef{{120, domain.VolumeUnitGm, 4000}}},
		{"Coca-Cola Soft Drink", "beverages", "Sunfeast", true, []variantDef{{750, domain.VolumeUnitMl, 4000}, {2000, domain.VolumeUnitMl, 9500}}},
	}

	for _, pd := range prodDefs {
		cat := cats[pd.cat]
		brand := brands[pd.brand]
		bid := brand.ID
		prod := domain.Product{
			CategoryID: cat.ID, BrandID: &bid, Name: pd.name,
			ImageURL:  sptr("https://picsum.photos/seed/" + cat.Slug + "/400"),
			IsTopItem: pd.top, Status: domain.StatusActive,
		}
		db.Create(&prod)
		db.Create(&domain.ProductImage{ProductID: prod.ID, ImageURL: *prod.ImageURL, DisplayOrder: 0})

		for vi, vd := range pd.variants {
			variant := domain.ProductVariant{
				ProductID: prod.ID, VolumeValue: vd.vol, VolumeUnit: vd.unit, DisplayOrder: vi,
			}
			db.Create(&variant)

			var prices []int64
			var minPrice int64
			var minPlatformID string
			for _, plat := range platforms {
				price := vd.price + seedPlatformOffset[plat.Code]
				// link + price per platform × city
				db.Create(&domain.ProductPlatformLink{
					VariantID: variant.ID, PlatformID: plat.ID,
					PlatformSKU: sptr(fmt.Sprintf("%s-%s", plat.Code, variant.ID[:8])),
					DeepLink:    sptr(fmt.Sprintf("https://%s.example/item/%s", plat.Code, variant.ID[:8])),
				})
				db.Create(&domain.PlatformPrice{
					VariantID: variant.ID, PlatformID: plat.ID, CityID: city.ID,
					PricePaise: price, MRPPaise: i64ptr(price + 5000), Available: true, Source: domain.PriceSourceManual,
				})
				prices = append(prices, price)
				if minPlatformID == "" || price < minPrice {
					minPrice = price
					minPlatformID = plat.ID
				}
			}
			avg, _ := util.AveragePaise(prices)
			mpid := minPlatformID
			db.Create(&domain.VariantPriceSummary{
				VariantID: variant.ID, CityID: city.ID,
				AvgPricePaise: i64ptr(avg), MinPricePaise: i64ptr(minPrice), MinPlatformID: &mpid,
				AvailablePlatformCount: len(prices),
			})
		}
	}

	// --- a home banner ---
	snacks := cats["beverages"]
	db.Create(&domain.Banner{
		ImageURL: "https://picsum.photos/seed/banner/800/300", TargetType: domain.BannerTargetCategory,
		TargetID: &snacks.ID, IsActive: true, DisplayOrder: 1,
	})

	zap.S().Infow("seed: demo data created", "city", city.Name, "platforms", len(platforms))
}
