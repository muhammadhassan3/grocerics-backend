package dto

// @Swagger:model ProductCategory
// @Property product_category_id: Unique identifier for the category
// @Property product_category_name: Display name of the category
// @Description Category a product belongs to.
type ProductCategory struct {
	// Unique identifier for the category
	ProductCategoryID string `json:"product_category_id"`
	// Display name of the category
	ProductCategoryName string `json:"product_category_name"`
}

// @Swagger:model Brand
// @Property product_brand_id: Unique identifier for the brand
// @Property product_brand_name: Display name of the brand
// @Description Brand a product belongs to.
type Brand struct {
	// Unique identifier for the brand
	ProductBrandID string `json:"product_brand_id"`
	// Display name of the brand
	ProductBrandName string `json:"product_brand_name"`
}

// @Swagger:model ProductSubCategory
// @Property product_sub_category_id: Unique identifier for the sub-category
// @Property product_sub_category_name: Display name of the sub-category
// @Description Sub-category a product belongs to. Zero-valued if the product has none.
type ProductSubCategory struct {
	ProductSubCategoryID   string `json:"product_sub_category_id"`
	ProductSubCategoryName string `json:"product_sub_category_name"`
}

// @Swagger:model ProductItem
// @Property product_id: Unique identifier for the product
// @Property image_url: URL of the product's display image
// @Property product_name: Display name of the product
// @Property product_description: Description of the product
// @Property product_category: Category the product belongs to
// @Property product_brand: Brand the product belongs to
// @Property is_top_item: Whether the product is flagged as a top/featured item
// @Property status: Status of the product
// @Property created_at: Creation timestamp, RFC3339
// @Description A single product row as shown in product listings.
type ProductItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Description of the product
	ProductDescription string `json:"product_description"`
	// Category the product belongs to
	ProductCategory ProductCategory `json:"product_category"`
	// Sub-category the product belongs to (zero-valued if none)
	ProductSubCategory ProductSubCategory `json:"product_sub_category"`
	// Brand the product belongs to
	ProductBrand Brand `json:"product_brand"`
	// Flat category id, echoing the write shape (same value as product_category.product_category_id)
	CategoryID string `json:"category_id"`
	// Flat sub-category id, echoing the write shape
	SubcategoryID string `json:"subcategory_id"`
	// Flat brand id, echoing the write shape
	BrandID string `json:"brand_id"`
	// Whether the product is flagged as a top/featured item
	IsTopItem bool `json:"is_top_item"`
	// Status of the product
	Status string `json:"status" enums:"active,disabled"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}
