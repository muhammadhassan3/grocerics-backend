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
	// Date the banner becomes visible, RFC3339
	StartDate string `json:"start_date"`
	// Date the banner stops being visible, RFC3339
	EndDate string `json:"end_date"`
	// URL of the banner image
	ImageURL string `json:"image_url"`
	// Whether the banner is currently enabled
	IsActive bool `json:"is_active"`
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
