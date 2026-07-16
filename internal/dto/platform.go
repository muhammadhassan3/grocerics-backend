package dto

import "grocerics-backend/internal/query"

// @Swagger:model PlatformItem
// @Description A delivery platform (Blinkit, Zepto, ...). qc_name maps it onto
// QuickCommerce; without one the platform cannot be searched or linked.
type PlatformItem struct {
	// Unique identifier for the platform
	PlatformID string `json:"platform_id"`
	// Short code used in requests, e.g. "blinkit"
	Code string `json:"code"`
	// Display name, e.g. "Blinkit"
	DisplayName string `json:"display_name"`
	// The platform's name on QuickCommerce, e.g. "BlinkIt". Empty = not mapped.
	QCName string `json:"qc_name,omitempty"`
	// True when qc_name is set, i.e. the platform can be searched/linked.
	Searchable bool `json:"searchable"`
	// URL of the platform's logo
	LogoURL string `json:"logo_url,omitempty"`
	// Fallback delivery ETA copy, e.g. "10 Mins"
	DeliveryETAText string `json:"delivery_eta_text,omitempty"`
	// Whether the platform is enabled
	Enabled bool `json:"enabled"`
	// Sort order in pickers
	DisplayOrder int `json:"display_order"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model Platforms
// @Description Paginated list of platforms.
type Platforms struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of platforms
	Platforms []PlatformItem `json:"platforms"`
}
