package dto

import "grocerics-backend/internal/query"

// @Swagger:model BrandItem
// @Property brand_id: Unique identifier for the brand
// @Property brand_name: Display name of the brand
// @Property image_url: URL of the brand's display image
// @Property status: Status of the brand
// @Property is_top_brand: Whether the brand is flagged as a top/featured brand
// @Property product_count: Number of products under this brand
// @Property created_at: Creation timestamp, RFC3339
// @Description A product brand as shown in brand listings.
type BrandItem struct {
	// Unique identifier for the brand
	BrandID string `json:"brand_id"`
	// Display name of the brand
	BrandName string `json:"brand_name"`
	// URL of the brand's display image
	ImageURL string `json:"image_url"`
	// Status of the brand
	Status string `json:"status" enums:"active,disabled"`
	// Whether the brand is flagged as a top/featured brand
	IsTopBrand bool `json:"is_top_brand"`
	// Curated sort order (lower shows first in the app)
	DisplayOrder int `json:"display_order"`
	// Number of products under this brand
	ProductCount int `json:"product_count"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model BrandList
// @Property meta: Pagination metadata
// @Property brands: Page of brands
// @Description Paginated list of brands.
type BrandList struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of brands
	Brands []BrandItem `json:"brands"`
}
