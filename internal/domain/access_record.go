package domain

import "time"

// used for admin anaylytics
type AccessRecord struct {
	BaseModel
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	Action    string    `gorm:"not null" json:"action"`
	IPAddress *string   `json:"ip_address,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
