package dto

import "grocerics-backend/internal/query"

// @Swagger:model InventoryManagementsStats
// @Property total_categories: Total number of categories
// @Property total_products: Total number of products
// @Property total_brands: Total number of brands
// @Property platforms: Total number of tracked delivery platforms
// @Description Headline stat cards shown on the inventory management dashboard.
type InventoryManagementsStats struct {
	// Total number of categories
	TotalCategories StatsItem `json:"total_categories"`
	// Total number of products
	TotalProducts StatsItem `json:"total_products"`
	// Total number of brands
	TotalBrands StatsItem `json:"total_brands"`
	// Total number of tracked delivery platforms
	Platforms StatsItem `json:"platforms"`
}

// @Swagger:model InventoryManagementsItem
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property image_url: URL of the product's display image
// @Property product_category: Name of the category the product belongs to
// @Property total_variants: Number of variants this product has
// @Property status: Status of the product
// @Property top_item: Whether the product is flagged as a top/featured item
// @Property stock_count: Units currently in stock across all variants
// @Description A single product row in the inventory management list.
type InventoryManagementsItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Name of the category the product belongs to
	ProductCategory string `json:"product_category"`
	// Number of variants this product has
	TotalVariants int `json:"total_variants"`
	// Status of the product
	Status string `json:"status" enums:"active,disabled"`
	// Whether the product is flagged as a top/featured item
	TopItem bool `json:"top_item"`
	// Units currently in stock across all variants
	StockCount int `json:"stock_count"`
}

// @Swagger:model InventoryManagements
// @Property meta: Pagination metadata
// @Property products: Page of inventory items
// @Description Paginated inventory management product list.
type InventoryManagements struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of inventory items
	Products []InventoryManagementsItem `json:"products"`
}
