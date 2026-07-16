package dto

import "grocerics-backend/internal/query"

// @Swagger:model TopDealsMobileItem
// @Property item_id: Unique identifier for the product
// @Property item_name: Display name of the product
// @Property category_name: Name of the category the product belongs to
// @Property image_url: URL of the product's display image
// @Property lowest_price: Lowest price for the product across tracked delivery platforms
// @Property cart_count: Number of times the product has been added to a cart
// @Description A single product row within the top deals list.
type TopDealsMobileItem struct {
	// Unique identifier for the product
	ItemID string `json:"item_id"`
	// Display name of the product
	ItemName string `json:"item_name"`
	// Name of the category the product belongs to
	CategoryName string `json:"category_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Lowest price for the product across tracked delivery platforms
	LowestPrice float64 `json:"lowest_price"`
	// Number of times the product has been added to a cart
	CartCount int `json:"cart_count"`
}

// @Swagger:model TopDealsMobileResponse
// @Property platforms: Delivery platforms considered when pricing the deals
// @Property results: Page of top deal products
// @Property meta: Pagination metadata
// @Description Envelope for the mobile app's top deals endpoint.
type TopDealsMobileResponse struct {
	// Delivery platforms considered when pricing the deals
	Platforms []PlatformSearchResult `json:"platforms"`
	// Page of top deal products
	Results []TopDealsMobileItem `json:"results"`
	// Pagination metadata
	Meta query.Meta `json:"meta"`
}
