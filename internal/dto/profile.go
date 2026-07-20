package dto

type MeDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email,omitempty"`
	Phone           string `json:"phone,omitempty"`
	CurrentCityID   string `json:"current_city_id,omitempty"`
	CurrentCityName string `json:"current_city_name,omitempty"`
	Onboarded       bool   `json:"onboarded"`
}

type OnboardingResponse struct {
	User    MeDTO      `json:"user"`
	Address AddressDTO `json:"address"`
}

type AddressDTO struct {
	ID          string   `json:"id"`
	Label       string   `json:"label,omitempty"`
	Line1       string   `json:"line1"`
	Line2       string   `json:"line2,omitempty"`
	Pincode     string   `json:"pincode"`
	CityID      string   `json:"city_id,omitempty"`
	CityName    string   `json:"city_name,omitempty"`
	Lat         *float64 `json:"lat,omitempty"`
	Lng         *float64 `json:"lng,omitempty"`
	IsDefault   bool     `json:"is_default"`
	Serviceable bool     `json:"serviceable"` // true if the pincode resolved to a serving city
}

// TODO: may update later, thisis just based of hte design
type NotificationPreferencesDTO struct {
	PriceAlerts  bool `json:"price_alerts"`
	Promotions   bool `json:"promotions"`
	OrderUpdates bool `json:"order_updates"`
}
