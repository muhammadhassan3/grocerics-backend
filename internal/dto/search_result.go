package dto

import "grocerics-backend/internal/query"

// @Swagger:model PlatformSearchResult
// @Property platform_id: Unique identifier for the delivery platform
// @Property platform_name: Display name of the delivery platform
// @Property platform_logo: URL of the delivery platform's logo
// @Description A delivery platform considered when searching or listing products.
type PlatformSearchResult struct {
	// Unique identifier for the delivery platform
	PlatformID string `json:"platform_id"`
	// Display name of the delivery platform
	PlatformName string `json:"platform_name"`
	// URL of the delivery platform's logo
	PlatformLogo string `json:"platform_logo"`
}

// @Swagger:model ResultItem
// @Property item_id: Unique identifier for the product
// @Property item_name: Display name of the product
// @Property brand_name: Display name of the product's brand
// @Property image_url: URL of the product's display image
// @Property lowest_price: Lowest price for the product across tracked delivery platforms
// @Property cart_count: Number of times the product has been added to a cart
// @Description A single product row within search results.
type ResultItem struct {
	// Unique identifier for the product
	ItemID string `json:"item_id"`
	// Display name of the product
	ItemName string `json:"item_name"`
	// Display name of the product's brand
	BrandName string `json:"brand_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Lowest price for the product across tracked delivery platforms
	LowestPrice float64 `json:"lowest_price"`
	// Number of times the product has been added to a cart
	CartCount int `json:"cart_count"`
}

// @Swagger:model SearchResultMobile
// @Property platforms: Delivery platforms considered when searching the products
// @Property results: Page of matching products
// @Property meta: Pagination metadata
// @Description Envelope for the mobile app's product search endpoint.
type SearchResultMobile struct {
	// Delivery platforms considered when searching the products
	Platforms []PlatformSearchResult `json:"platforms"`
	// Page of matching products
	Results []ResultItem `json:"results"`
	// Pagination metadata
	Meta query.Meta `json:"meta"`
}
