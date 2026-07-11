package domain

type Wishlist struct {
	BaseModel
	UserID    string `gorm:"type:uuid;not null" json:"user_id"`
	VariantID string `gorm:"type:uuid;not null" json:"variant_id"`
	Timestamps
	SoftDelete
}
