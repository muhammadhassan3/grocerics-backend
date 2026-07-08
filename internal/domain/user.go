package domain

// User is the auth principal. company_id is nullable (admins belong to no
// company per decisions.md). The DB column for the hash stays `password`
// for backward compat with the auth dev's migration; the Go field is named
// honestly. PasswordHash never serializes (json:"-").
type User struct {
	BaseModel
	Name         string     `gorm:"not null" json:"name"`
	Email        string     `gorm:"not null" json:"email"`
	PasswordHash string     `gorm:"not null;column:password" json:"-"`
	Role         Role       `gorm:"type:varchar;not null" json:"role"`
	Status       UserStatus `gorm:"type:varchar;not null;default:'active'" json:"status"`
	Timestamps
	SoftDelete
}
