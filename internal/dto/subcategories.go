package dto

import "grocerics-backend/internal/query"

// @Swagger:model SubCategory
// @Property sub_category_id: Unique identifier for the subcategory
// @Property sub_category_name: Display name of the subcategory
// @Property image_url: URL of the subcategory's display image
// @Property status: Status of the subcategory
// @Property is_top_sub_category: Whether the subcategory is flagged as a top/featured subcategory
// @Property category_id: Identifier of the parent category
// @Property created_at: Creation timestamp, RFC3339
// @Description A subcategory nested under a product category.
type SubCategory struct {
	// Unique identifier for the subcategory
	SubCategoryID string `json:"sub_category_id"`
	// Display name of the subcategory
	SubCategoryName string `json:"sub_category_name"`
	// URL of the subcategory's display image
	ImageURL string `json:"image_url"`
	// Status of the subcategory
	Status string `json:"status" enums:"active,disabled"`
	// Whether the subcategory is flagged as a top/featured subcategory
	IsTopSubCategory bool `json:"is_top_sub_category"`
	// Identifier of the parent category
	CategoryID string `json:"category_id"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model SubCategories
// @Property meta: Pagination metadata
// @Property sub_categories: Page of subcategories
// @Description Paginated list of subcategories.
type SubCategories struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of subcategories
	SubCategories []SubCategory `json:"sub_categories"`
}
