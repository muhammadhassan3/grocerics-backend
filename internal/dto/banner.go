package dto

import "grocerics-backend/internal/query"

// @Swagger:model BannerItem
// @Property banner_id: Unique identifier for the banner
// @Property start_date: Date the banner becomes visible, RFC3339
// @Property end_date: Date the banner stops being visible, RFC3339
// @Property image_url: URL of the banner image
// @Property is_active: Whether the banner is currently enabled
// @Property created_at: Creation timestamp, RFC3339
// @Description A promotional banner shown on the storefront.
type BannerItem struct {
	// Unique identifier for the banner
	BannerID string `json:"banner_id"`
	// Admin-facing name, so banners are distinguishable in the grid
	Title string `json:"title"`
	// Date the banner becomes visible, RFC3339
	StartDate string `json:"start_date"`
	// Date the banner stops being visible, RFC3339
	EndDate string `json:"end_date"`
	// URL of the banner image
	ImageURL string `json:"image_url"`
	// Manual on/off switch. Can only turn a banner OFF — it never forces one
	// live outside its date window.
	IsActive bool `json:"is_active"`
	// Derived, read-only: whether the banner is actually showing right now
	// (is_active AND inside the date window). This is what the grid should display.
	IsLive bool `json:"is_live"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model Banners
// @Property meta: Pagination metadata
// @Property banners: Page of banners
// @Description Paginated list of banners.
type Banners struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of banners
	Banners []BannerItem `json:"banners"`
}
