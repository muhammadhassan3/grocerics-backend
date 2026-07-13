package dto

// @Swagger:model PlatformMobilePrice
// @Property platform_id: Unique identifier for the delivery platform
// @Property platform_name: Display name of the delivery platform
// @Property platform_image_url: URL of the delivery platform's logo
// @Property platform_price: Price of the product on this platform, formatted for display
// @Description Price of a product on a single delivery platform.
type PlatformMobilePrice struct {
	// Unique identifier for the delivery platform
	PlatformID string `json:"platform_id"`
	// Display name of the delivery platform
	PlatformName string `json:"platform_name"`
	// URL of the delivery platform's logo
	PlatformImageURL string `json:"platform_image_url"`
	// Price of the product on this platform, formatted for display
	PlatformPrice string `json:"platform_price"`
}

// @Swagger:model ProductItemDetail
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property quantity: Number of units of this product
// @Property volume_formatted: Human-readable volume/weight of the product, e.g. "500 gm"
// @Property average_price: Average price across tracked delivery platforms, formatted for display
// @Description Summary of a product used within order/detail contexts.
type ProductItemDetail struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// Number of units of this product
	Quantity int `json:"quantity"`
	// Human-readable volume/weight of the product, e.g. "500 gm"
	VolumeFormatted string `json:"volume_formatted"`
	// Average price across tracked delivery platforms, formatted for display
	AveragePrice string `json:"average_price"`
}

// @Swagger:model SimilarProducts
// @Property item_id: Unique identifier for the product
// @Property item_name: Display name of the product
// @Property brand_name: Display name of the product's brand
// @Property image_url: URL of the product's display image
// @Property lowest_price: Lowest price for the product across tracked delivery platforms
// @Property cart_count: Number of times the product has been added to a cart
// @Description A product recommended as similar to the one being viewed.
type SimilarProducts struct {
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

// @Swagger:model ProductDetailMobileResponse
// @Property images: Gallery of image URLs for the product
// @Property is_wishlist: Whether the product is in the current user's wishlist
// @Property prices: Per-platform prices for the product
// @Property product_description: Description of the product
// @Property similar_products: Products recommended as similar to this one
// @Description Envelope for the mobile app's product detail endpoint.
type ProductDetailMobileResponse struct {
	// Gallery of image URLs for the product
	Images []string `json:"images"`
	// Whether the product is in the current user's wishlist
	IsWishlist bool `json:"is_wishlist"`
	// Per-platform prices for the product
	Prices []PlatformMobilePrice `json:"prices"`
	// Description of the product
	ProductDescription string `json:"product_description"`
	// Products recommended as similar to this one
	SimilarProducts []SimilarProducts `json:"similar_products"`
}
