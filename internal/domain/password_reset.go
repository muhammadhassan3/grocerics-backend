package domain

import "time"

// PasswordReset is a single-use token. Same hash discipline as RefreshToken.
// UsedAt non-nil = spent.
type PasswordReset struct {
	BaseModel
	UserID    string     `gorm:"type:uuid;not null" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(64);not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
