package domain

import "time"

// RefreshToken is stored as a SHA-256 hash of the JWT — never plaintext
// (see decisions.md). No SoftDelete: tokens use revoked_at + expires_at
// for their lifecycle.
type RefreshToken struct {
	BaseModel
	UserID    string     `gorm:"type:uuid;not null" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(64);not null;column:token_hash" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

// IsActive reports whether the token can still be used to mint access tokens.
func (t *RefreshToken) IsActive(now time.Time) bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(now)
}
