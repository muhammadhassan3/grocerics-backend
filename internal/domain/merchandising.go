package domain

import "time"

type Banner struct {
	BaseModel
	ImageURL     string           `gorm:"not null" json:"image_url"`
	TargetType   BannerTargetType `gorm:"type:varchar;not null;default:'none'" json:"target_type"`
	TargetID     *string          `gorm:"type:uuid" json:"target_id,omitempty"`
	TargetURL    *string          `json:"target_url,omitempty"`
	StartDate    *time.Time       `json:"start_date,omitempty"`
	EndDate      *time.Time       `json:"end_date,omitempty"`
	IsActive     bool             `gorm:"not null;default:true" json:"is_active"`
	DisplayOrder int              `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

// powering the admin dashboards "total searches" and "top searched products".
type SearchEvent struct {
	BaseModel
	UserID          *string   `gorm:"type:uuid" json:"user_id,omitempty"`
	Query           string    `gorm:"not null" json:"query"`
	ResultProductID *string   `gorm:"type:uuid" json:"result_product_id,omitempty"`
	CityID          *string   `gorm:"type:uuid" json:"city_id,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}
