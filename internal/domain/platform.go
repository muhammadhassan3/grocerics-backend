package domain

type Platform struct {
	BaseModel
	Code             string  `gorm:"not null" json:"code"`
	DisplayName      string  `gorm:"not null" json:"display_name"`
	LogoURL          *string `json:"logo_url,omitempty"`
	DeepLinkTemplate *string `json:"deep_link_template,omitempty"`
	DeliveryETAText  *string `json:"delivery_eta_text,omitempty"`
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
