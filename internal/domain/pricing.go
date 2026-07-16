package domain

import "time"

// PlatformPrice is the price + availability of one variant on one platform in
// one city.
type PlatformPrice struct {
	BaseModel
	VariantID     string      `gorm:"type:uuid;not null" json:"variant_id"`
	PlatformID    string      `gorm:"type:uuid;not null" json:"platform_id"`
	CityID        string      `gorm:"type:uuid;not null" json:"city_id"`
	PricePaise    int64       `gorm:"not null" json:"price_paise"`
	MRPPaise      *int64      `json:"mrp_paise,omitempty"` // for discount display
	Available     bool        `gorm:"not null" json:"available"`
	Inventory     *int        `json:"inventory,omitempty"`
	Source        PriceSource `gorm:"type:varchar;not null;default:'api'" json:"source"`
	LastUpdatedAt time.Time   `gorm:"not null;autoUpdateTime" json:"last_updated_at"` // freshness
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

// price of variant and its details
type VariantPriceSummary struct {
	VariantID              string    `gorm:"type:uuid;primaryKey" json:"variant_id"`
	CityID                 string    `gorm:"type:uuid;primaryKey" json:"city_id"`
	AvgPricePaise          *int64    `json:"avg_price_paise,omitempty"`
	MinPricePaise          *int64    `json:"min_price_paise,omitempty"`
	MinPlatformID          *string   `gorm:"type:uuid" json:"min_platform_id,omitempty"`
	AvailablePlatformCount int       `gorm:"not null;default:0" json:"available_platform_count"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
