package dto

// grid card used by Home rails, PLP, Search and Top Deals.
type ProductCardDTO struct {
	ProductID        string    `json:"product_id"`
	DefaultVariantID string    `json:"default_variant_id,omitempty"`
	Name             string    `json:"name"`
	BrandName        string    `json:"brand_name,omitempty"`
	ImageURL         string    `json:"image_url,omitempty"`
	StartingPrice    *MoneyDTO `json:"starting_price,omitempty"`
}

// a delivery platform (Home "Top Stores" + filter chips).
type PlatformDTO struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	LogoURL         string `json:"logo_url,omitempty"`
	DeliveryETAText string `json:"delivery_eta_text,omitempty"`
}

//one platforms price on the PDP.

type PlatformPriceChipDTO struct {
	PlatformCode string    `json:"platform_code"`
	PlatformName string    `json:"platform_name"`
	LogoURL      string    `json:"logo_url,omitempty"`
	Price        MoneyDTO  `json:"price"`
	MRP          *MoneyDTO `json:"mrp,omitempty"`
	Available    bool      `json:"available"`
	DeepLink     string    `json:"deep_link,omitempty"`
}

type HomeResponse struct {
	Banners       []BannerCardDTO   `json:"banners"`
	TopStores     []PlatformDTO     `json:"top_stores"`
	Categories    []CategoryCardDTO `json:"categories"`
	TrendingItems []ProductCardDTO  `json:"trending_items"`
}

type BannerCardDTO struct {
	ImageURL   string `json:"image_url"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id,omitempty"`
	TargetURL  string `json:"target_url,omitempty"`
}

type CategoryCardDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ImageURL string `json:"image_url,omitempty"`
}

type CityDTO struct {
	ID   string `json:"city_id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductDetailDTO struct {
	ProductID   string             `json:"product_id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	BrandName   string             `json:"brand_name,omitempty"`
	CategoryID  string             `json:"category_id"`
	Images      []string           `json:"images"`
	Variants    []VariantDetailDTO `json:"variants"`
	Similar     []ProductCardDTO   `json:"similar"`
}

type VariantDetailDTO struct {
	VariantID      string                 `json:"variant_id"`
	PackLabel      string                 `json:"pack_label"`           // "500 gm"
	UnitPrice      string                 `json:"unit_price,omitempty"` // "₹72/100gm"
	AveragePrice   *MoneyDTO              `json:"average_price,omitempty"`
	PlatformPrices []PlatformPriceChipDTO `json:"platform_prices"`
}

type CartResponse struct {
	CartID    string                 `json:"cart_id"`
	Items     []CartLineDTO          `json:"items"`
	Platforms []PlatformBreakdownDTO `json:"platforms"`
}

type CartLineDTO struct {
	ItemID       string    `json:"item_id"`
	VariantID    string    `json:"variant_id"`
	ProductName  string    `json:"product_name"`
	PackLabel    string    `json:"pack_label"`
	Quantity     int       `json:"quantity"`
	AveragePrice *MoneyDTO `json:"average_price,omitempty"`
}

// which items are available, the
// item split, and totals. Reused by cart and wishlist.
type PlatformBreakdownDTO struct {
	PlatformCode       string   `json:"platform_code"`
	PlatformName       string   `json:"platform_name"`
	LogoURL            string   `json:"logo_url,omitempty"`
	DeliveryETAMinutes *int     `json:"delivery_eta_minutes,omitempty"`
	AvailableItemIDs   []string `json:"available_item_ids"`
	UnavailableItemIDs []string `json:"unavailable_item_ids"`
	ItemTotal          MoneyDTO `json:"item_total"`
	AvailableTotal     MoneyDTO `json:"available_total"`
	IsCheapest         bool     `json:"is_cheapest"`
}
