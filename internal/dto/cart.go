package dto

// @Swagger:model CartItem
// @Property product_id: Unique identifier for the product
// @Property product_name: Display name of the product
// @Property image_url: URL of the product's display image
// @Property quantity: Number of units of this product in the cart
// @Property volume_formatted: Human-readable volume/weight of the product, e.g. "500 gm"
// @Property average_price: Average price across tracked delivery platforms, formatted for display
// @Description A single product line item within a user's cart.
type CartItem struct {
	// Unique identifier for the product
	ProductID string `json:"product_id"`
	// Display name of the product
	ProductName string `json:"product_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Number of units of this product in the cart
	Quantity int `json:"quantity"`
	// Human-readable volume/weight of the product, e.g. "500 gm"
	VolumeFormatted string `json:"volume_formatted"`
	// Average price across tracked delivery platforms, formatted for display
	AveragePrice string `json:"average_price"`
}

// @Swagger:model CartMobile
// @Property platforms: Delivery platforms considered when pricing the cart
// @Property eta: Estimated delivery time for the cart, e.g. "10 mins"
// @Property available: Cart items that are currently in stock
// @Property unavailable: Cart items that are currently out of stock
// @Description Mobile cart response, split into available and unavailable items.
type CartMobile struct {
	// Delivery platforms considered when pricing the cart
	Platfoms []PlatformSearchResult `json:"platforms"`
	// Estimated delivery time for the cart, e.g. "10 mins"
	Eta string `json:"eta"`
	// Cart items that are currently in stock
	Available []CartItem `json:"available"`
	// Cart items that are currently out of stock
	Unavailable []CartItem `json:"unavailable"`
}
