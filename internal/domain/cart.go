package domain

type Cart struct {
	BaseModel
	UserID string `gorm:"type:uuid;not null" json:"user_id"`
	Timestamps
	SoftDelete
}

type CartItem struct {
	BaseModel
	CartID    string `gorm:"type:uuid;not null" json:"cart_id"`
	VariantID string `gorm:"type:uuid;not null" json:"variant_id"`
	Quantity  int    `gorm:"not null;default:1" json:"quantity"`
	Timestamps
	SoftDelete
}
