package domain

import "time"

type PhoneOTP struct {
	BaseModel
	UserID       *string    `gorm:"type:uuid" json:"user_id,omitempty"`
	Phone        string     `gorm:"not null" json:"phone"`
	OTPCodeHash  string     `gorm:"not null" json:"-"`
	Purpose      OTPPurpose `gorm:"type:varchar;not null" json:"purpose"`
	AttemptCount int        `gorm:"not null;default:0" json:"attempt_count"`
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	Consumed     bool       `gorm:"not null;default:false" json:"consumed"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (o *PhoneOTP) IsUsable(now time.Time) bool {
	return !o.Consumed && o.VerifiedAt == nil && o.ExpiresAt.After(now)
}
