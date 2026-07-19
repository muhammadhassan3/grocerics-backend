package domain

// User is the mobile client. Auth is phone + OTP only; email/password/role were
// removed in migration 00016 (admins are a separate table now). `phone` is the
// login identity — NOT NULL and partial-unique (WHERE deleted_at IS NULL).
type User struct {
	BaseModel
	Name          string     `gorm:"not null" json:"name"`
	Phone         string     `gorm:"not null" json:"phone"`
	Status        UserStatus `gorm:"type:varchar;not null;default:'active'" json:"status"`
	CurrentCityID *string    `gorm:"type:uuid" json:"current_city_id,omitempty"`
	Timestamps
	SoftDelete
}
