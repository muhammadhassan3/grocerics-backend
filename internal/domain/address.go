package domain

type UserAddress struct {
	BaseModel
	UserID    string   `gorm:"type:uuid;not null" json:"user_id"`
	Label     *string  `json:"label,omitempty"`
	Line1     string   `gorm:"not null" json:"line1"`
	Line2     *string  `json:"line2,omitempty"`
	Pincode   string   `gorm:"not null" json:"pincode"`
	CityID    *string  `gorm:"type:uuid" json:"city_id,omitempty"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
	IsDefault bool     `gorm:"not null;default:false" json:"is_default"`
	Timestamps
	SoftDelete
}
