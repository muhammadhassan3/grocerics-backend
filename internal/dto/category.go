package dto

import "grocerics-backend/internal/query"

// @Swagger:model Category
// @Property category_id: Unique identifier for the category
// @Property category_name: Display name of the category
// @Property image_url: URL of the category's display image
// @Property sub_category_count: Number of subcategories nested under this category
// @Property status: Status of the category
// @Property is_top_category: Whether the category is flagged as a top/featured category
// @Property created_at: Creation timestamp, RFC3339
// @Description A top-level product category.
type Category struct {
	// Unique identifier for the category
	CategoryID string `json:"category_id"`
	// Display name of the category
	CategoryName string `json:"category_name"`
	// URL of the category's display image
	ImageURL string `json:"image_url"`
	// Number of subcategories nested under this category
	SubCategoryCount int `json:"sub_category_count"`
	// Status of the category
	Status string `json:"status" enums:"active,disabled"`
	// Whether the category is flagged as a top/featured category
	IsTopCategory bool `json:"is_top_category"`
	// Curated sort order (lower shows first in the app)
	DisplayOrder int `json:"display_order"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model Categories
// @Property meta: Pagination metadata
// @Property categories: Page of categories
// @Description Paginated list of categories.
type Categories struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of categories
	Categories []Category `json:"categories"`
}
