package dto

// @Swagger:model BannerMobileItem
// @Property banner_id: Unique identifier for the banner
// @Property image_url: URL of the banner image
// @Property deeplink: App deeplink to navigate to when the banner is tapped
// @Description A promotional banner shown on the mobile dashboard.
type BannerMobileItem struct {
	// Unique identifier for the banner
	BannerID string `json:"banner_id"`
	// URL of the banner image
	ImageURL string `json:"image_url"`
	// App deeplink to navigate to when the banner is tapped
	Deeplink string `json:"deeplink"`
}

// @Swagger:model TopStoreMobileItem
// @Property store_id: Unique identifier for the store
// @Property image_url: URL of the store's display image
// @Property store_name: Display name of the store
// @Property eta: Estimated delivery time for the store, e.g. "10 mins"
// @Description A featured store shown on the mobile dashboard.
type TopStoreMobileItem struct {
	// Unique identifier for the store
	StoreID string `json:"store_id"`
	// URL of the store's display image
	ImageURL string `json:"image_url"`
	// Display name of the store
	StoreName string `json:"store_name"`
	// Estimated delivery time for the store, e.g. "10 mins"
	ETA string `json:"eta"`
}

// @Swagger:model TrendingCategoryMobileItem
// @Property category_id: Unique identifier for the category
// @Property category_name: Display name of the category
// @Property image_url: URL of the category's display image
// @Description A trending product category shown on the mobile dashboard.
type TrendingCategoryMobileItem struct {
	// Unique identifier for the category
	CategoryID string `json:"category_id"`
	// Display name of the category
	CategoryName string `json:"category_name"`
	// URL of the category's display image
	ImageURL string `json:"image_url"`
}

// @Swagger:model TrendingItemMobileItem
// @Property item_id: Unique identifier for the product
// @Property item_name: Display name of the product
// @Property image_url: URL of the product's display image
// @Property lowest_price: Lowest price for the product across tracked delivery platforms
// @Description A trending product shown on the mobile dashboard.
type TrendingItemMobileItem struct {
	// Unique identifier for the product
	ItemID string `json:"item_id"`
	// Display name of the product
	ItemName string `json:"item_name"`
	// URL of the product's display image
	ImageURL string `json:"image_url"`
	// Lowest price for the product across tracked delivery platforms
	LowestPrice float64 `json:"lowest_price"`
}

// @Swagger:model DashboardMobile
// @Property banners: Promotional banners for the mobile dashboard
// @Property top_stores: Featured stores for the mobile dashboard
// @Property trending_categories: Trending product categories for the mobile dashboard
// @Property trending_items: Trending products for the mobile dashboard
// @Description Envelope for the mobile app's home dashboard endpoint.
type DashboardMobile struct {
	// Promotional banners for the mobile dashboard
	Banners []BannerMobileItem `json:"banners"`
	// Featured stores for the mobile dashboard
	TopStores []TopStoreMobileItem `json:"top_stores"`
	// Trending product categories for the mobile dashboard
	TrendingCategories []TrendingCategoryMobileItem `json:"trending_categories"`
	// Trending products for the mobile dashboard
	TrendingItems []TrendingItemMobileItem `json:"trending_items"`
}
