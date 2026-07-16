package domain

type FcmToken struct {
	BaseModel
	UserID   string         `gorm:"type:uuid;not null" json:"user_id"`
	Token    string         `gorm:"not null" json:"token"`
	Platform DevicePlatform `gorm:"type:varchar;not null;default:'android'" json:"platform"`
	Timestamps
	SoftDelete
}

type NotificationPreference struct {
	BaseModel
	UserID       string `gorm:"type:uuid;not null" json:"user_id"`
	PriceAlerts  bool   `gorm:"not null" json:"price_alerts"`
	Promotions   bool   `gorm:"not null" json:"promotions"`
	OrderUpdates bool   `gorm:"not null" json:"order_updates"`
	Timestamps
	SoftDelete
}
