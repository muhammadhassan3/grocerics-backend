package dto

import "grocerics-backend/internal/query"

// @Swagger:model PlatformPriceItem
// @Description One active platform's live price for a product. Money is integer paise (3800 = ₹38.00).
type PlatformPriceItem struct {
	// Platform identifier
	PlatformID string `json:"platform_id"`
	// Platform code, e.g. "blinkit"
	PlatformCode string `json:"platform_code"`
	// Platform display name, e.g. "Blinkit"
	PlatformName string `json:"platform_name"`
	// Selling price in paise (0 if this platform has no price for the product)
	PricePaise int64 `json:"price_paise"`
	// MRP in paise (0 when unknown)
	MRPPaise int64 `json:"mrp_paise"`
	// Whether the product is available on this platform
	Available bool `json:"available"`
}

// @Swagger:model ProductPrice
// @Description A product with its live price across every active platform.
type ProductPrice struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Name of the category the product belongs to
	ProductCategory string `json:"product_category"`
	// One entry per active platform (dynamic — not a fixed set)
	PlatformPrices []PlatformPriceItem `json:"platform_prices"`
}

// @Swagger:model LivePrice
// @Description Paginated list of live product prices.
type LivePrice struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of product prices
	Products []ProductPrice `json:"products"`
}
