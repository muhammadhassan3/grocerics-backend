package dto

import "grocerics-backend/internal/query"

// @Swagger:model TopSearchProductItem
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property product_category: Name of the category the product belongs to
// @Property search_count: Number of times the product was searched for
// @Description A product ranked by search volume.
type TopSearchProductItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Name of the category the product belongs to
	ProductCategory string `json:"product_category"`
	// Number of times the product was searched for
	SearchCount int `json:"search_count"`
}

// @Swagger:model TopSearchProduct
// @Property products: Page of ranked products
// @Property meta: Pagination metadata
// @Description Paginated list of top-searched products.
type TopSearchProduct struct {
	// Page of ranked products
	Products []TopSearchProductItem `json:"products"`
	// Pagination metadata
	Meta query.Meta `json:"meta"`
}
