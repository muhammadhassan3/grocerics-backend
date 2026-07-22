package dto

import "grocerics-backend/internal/query"

// @Swagger:model CityItem
// @Description A serviceable city. The lat/lng + default_pincode are the location
// anchor admin QuickCommerce calls are made from.
type CityItem struct {
	// Unique identifier for the city
	CityID string `json:"city_id"`
	// Display name of the city
	Name string `json:"name"`
	// URL-safe name
	Slug string `json:"slug"`
	// State the city sits in
	State string `json:"state,omitempty"`
	// Latitude of the city's QuickCommerce location anchor
	Lat *float64 `json:"lat,omitempty"`
	// Longitude of the city's QuickCommerce location anchor
	Lng *float64 `json:"lng,omitempty"`
	// Pincode used as the default QuickCommerce location anchor
	DefaultPincode string `json:"default_pincode,omitempty"`
	// Serviceability status: "active" | "disabled"
	Status string `json:"status"`
}

// @Swagger:model Cities
// @Description Paginated list of cities.
type Cities struct {
	// Pagination metadata
	Meta query.Meta `json:"meta"`
	// Page of cities
	Cities []CityItem `json:"cities"`
}
