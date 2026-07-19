package domain

import "time"

type Admin struct {
	BaseModel
	Name         string     `gorm:"not null" json:"name"`
	Email        string     `gorm:"not null" json:"email"`
	PasswordHash string     `gorm:"not null;column:password_hash" json:"-"`
	Status       UserStatus `gorm:"type:varchar;not null;default:'active'" json:"status"`
	Timestamps
	SoftDelete
}

type AdminRefreshToken struct {
	BaseModel
	AdminID   string     `gorm:"type:uuid;not null" json:"admin_id"`
	TokenHash string     `gorm:"type:varchar(64);not null;column:token_hash" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (t AdminRefreshToken) IsActive(now time.Time) bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(now)
}

type AdminPasswordReset struct {
	BaseModel
	AdminID   string     `gorm:"type:uuid;not null" json:"admin_id"`
	TokenHash string     `gorm:"type:varchar(64);not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
