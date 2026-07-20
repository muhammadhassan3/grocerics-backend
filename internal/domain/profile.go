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
	UserID     string `gorm:"type:uuid;not null" json:"user_id"`
	Muted      bool   `gorm:"not null;default:false" json:"muted"`
	Promotions bool   `gorm:"not null" json:"promotions"`
	Deals      bool   `gorm:"not null" json:"deals"`
	Timestamps
	SoftDelete
}
