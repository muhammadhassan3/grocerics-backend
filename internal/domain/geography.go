package domain

type City struct {
	BaseModel
	Name         string  `gorm:"not null" json:"name"`
	Slug         string  `gorm:"not null" json:"slug"`
	State        *string `json:"state,omitempty"`
	Enabled      bool    `gorm:"not null;default:true" json:"enabled"`
	DisplayOrder int     `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type Pincode struct {
	BaseModel
	Pincode     string `gorm:"not null" json:"pincode"`
	CityID      string `gorm:"type:uuid;not null" json:"city_id"`
	Serviceable bool   `gorm:"not null;default:true" json:"serviceable"`
	Timestamps
	SoftDelete
}
