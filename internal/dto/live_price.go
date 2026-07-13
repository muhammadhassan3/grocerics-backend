package dto

import "grocerics-backend/internal/query"

// @Swagger:model PlatformPrice
// @Property blinkit_price: Current price on Blinkit
// @Property zepto_price: Current price on Zepto
// @Property instamart_price: Current price on Instamart
// @Description Live price of a product across tracked delivery platforms.
type PlatformPrice struct {
	// Current price on Blinkit
	BlinkitPrice float64 `json:"blinkit_price"`
	// Current price on Zepto
	ZeptoPrice float64 `json:"zepto_price"`
	// Current price on Instamart
	InstamartPrice float64 `json:"instamart_price"`
}

// @Swagger:model ProductPrice
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property product_category: Name of the category the product belongs to
// @Property platform_price: Per-platform live prices
// @Description A product with its live price across platforms.
type ProductPrice struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Name of the category the product belongs to
	ProductCategory string `json:"product_category"`
	// Per-platform live prices
	Proces PlatformPrice `json:"platform_price"`
}

// @Swagger:model LivePrice
// @Property meta: Pagination metadata
// @Property products: Page of product prices
// @Description Paginated list of live product prices.
type LivePrice struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of product prices
	Products []ProductPrice `json:"products"`
}
