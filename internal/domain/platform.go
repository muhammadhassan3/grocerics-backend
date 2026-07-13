package domain

import "time"

type Platform struct {
	BaseModel
	Code             string  `gorm:"not null" json:"code"`
	DisplayName      string  `gorm:"not null" json:"display_name"`
	LogoURL          *string `json:"logo_url,omitempty"`
	DeepLinkTemplate *string `json:"deep_link_template,omitempty"`
	DeliveryETAText  *string `json:"delivery_eta_text,omitempty"` // default ETA fallback; real ETA is per-pincode (platform_delivery_etas)
	Enabled          bool    `gorm:"not null;default:true" json:"enabled"`
	DisplayOrder     int     `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type ProductPlatformLink struct {
	BaseModel
	VariantID   string  `gorm:"type:uuid;not null" json:"variant_id"`
	PlatformID  string  `gorm:"type:uuid;not null" json:"platform_id"`
	PlatformSKU *string `json:"platform_sku,omitempty"`
	ProductURL  *string `json:"product_url,omitempty"`
	DeepLink    *string `json:"deep_link,omitempty"`
	Timestamps
	SoftDelete
}

// Prices are city-level, the delivery ETA is based on pincode, because the
// same city can be 10 vs 25 minutes depending on the area
type PlatformDeliveryETA struct {
	BaseModel
	PlatformID    string    `gorm:"type:uuid;not null" json:"platform_id"`
	Pincode       string    `gorm:"not null" json:"pincode"`
	ETAMinutes    *int      `json:"eta_minutes,omitempty"`
	Serviceable   bool      `gorm:"not null;default:true" json:"serviceable"` // platform may not deliver to this pincode at all
	LastUpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"last_updated_at"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PlatformDeliveryETA) TableName() string { return "platform_delivery_etas" }
