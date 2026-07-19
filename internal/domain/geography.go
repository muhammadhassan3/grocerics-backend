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
	Timestamps
	SoftDelete
}

