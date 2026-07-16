package domain

type City struct {
	BaseModel
	Name           string   `gorm:"not null" json:"name"`
	Slug           string   `gorm:"not null" json:"slug"`
	State          *string  `json:"state,omitempty"`
	Lat            *float64 `json:"lat,omitempty"`
	Lng            *float64 `json:"lng,omitempty"`
	DefaultPincode *string  `json:"default_pincode,omitempty"`
	Enabled        bool     `gorm:"not null" json:"enabled"`
	DisplayOrder   int      `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type Pincode struct {
	BaseModel
	Pincode     string   `gorm:"not null" json:"pincode"`
	CityID      string   `gorm:"type:uuid;not null" json:"city_id"`
	Lat         *float64 `json:"lat,omitempty"`
	Lng         *float64 `json:"lng,omitempty"`
	Serviceable bool     `gorm:"not null" json:"serviceable"`
	Timestamps
	SoftDelete
}
