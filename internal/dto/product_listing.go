package dto

import "grocerics-backend/internal/query"

// @Swagger:model ProductListingItem
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property brand_name: Display name of the product's brand
// @Property image_url: URL of the product's display image
// @Property average_price: Average price across tracked delivery platforms, formatted for display
// @Property cart_count: Number of times the product has been added to a cart
// @Description A single product row within a product listing.
type ProductListingItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Display name of the product's brand
	BrandName string `json:"brand_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Average price across tracked delivery platforms, formatted for display
	AveragePrice string `json:"average_price"`
	// Number of times the product has been added to a cart
	CartCount int `json:"cart_count"`
}

// @Swagger:model ProductListingMobile
// @Property platforms: Delivery platforms considered when listing the products
// @Property results: Page of listed products
// @Property meta: Pagination metadata
// @Description Envelope for the mobile app's product listing endpoint.
type ProductListingMobile struct {
	// Delivery platforms considered when listing the products
	Platforms []PlatformSearchResult `json:"platforms"`
	// Page of listed products
	Results []ProductListingItem `json:"results"`
	// Pagination metadata
	Meta query.Meta `json:"meta"`
}
