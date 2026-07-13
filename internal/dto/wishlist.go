package dto

// @Swagger:model WishlistItem
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property image_url: URL of the product's display image
// @Property quantity: Number of units of this product
// @Property volume_formatted: Human-readable volume/weight of the product, e.g. "500 gm"
// @Property average_price: Average price across tracked delivery platforms, formatted for display
// @Description A single product line item within a user's wishlist.
type WishlistItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Number of units of this product
	Quantity int `json:"quantity"`
	// Human-readable volume/weight of the product, e.g. "500 gm"
	VolumeFormatted string `json:"volume_formatted"`
	// Average price across tracked delivery platforms, formatted for display
	AveragePrice string `json:"average_price"`
}

// @Swagger:model WishlistMobile
// @Property platforms: Delivery platforms considered when pricing the wishlist
// @Property available: Wishlist items that are currently in stock
// @Property unavailable: Wishlist items that are currently out of stock
// @Description Mobile wishlist response, split into available and unavailable items.
type WishlistMobile struct {
	// Delivery platforms considered when pricing the wishlist
	Platforms []PlatformSearchResult `json:"platforms"`
	// Wishlist items that are currently in stock
	Available []WishlistItem `json:"available"`
	// Wishlist items that are currently out of stock
	Unavailable []WishlistItem `json:"unavailable"`
}
