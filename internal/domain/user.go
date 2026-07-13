package domain

// User is the auth principal. Two login paths will coexist on this table:
// email + password (admin / web, built) and phone + OTP (mobile app — the
// `phone` column and phone_otps table land now; the OTP endpoints are a
// later pass, at which point `email`/`password` become nullable).
//
// The DB column for the hash stays `password` for backward compat with the
// original auth migration; the Go field is named honestly and never
// serializes (json:"-"). `phone` is nullable + partial-unique (WHERE
// deleted_at IS NULL).
type User struct {
	BaseModel
	Name          string     `gorm:"not null" json:"name"`
	Email         string     `gorm:"not null" json:"email"`
	PasswordHash  string     `gorm:"not null;column:password" json:"-"`
	Phone         *string    `json:"phone,omitempty"`
	Role          Role       `gorm:"type:varchar;not null" json:"role"`
	Status        UserStatus `gorm:"type:varchar;not null;default:'active'" json:"status"`
	CurrentCityID *string    `gorm:"type:uuid" json:"current_city_id,omitempty"`
	Timestamps
	SoftDelete
}
