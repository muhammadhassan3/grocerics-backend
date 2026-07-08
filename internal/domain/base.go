// Package domain contains domain models used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded by every persisted model that has a UUID primary key.
type BaseModel struct {
	ID string `gorm:"type:uuid;primaryKey" json:"id"`
}

// BeforeCreate fires automatically before any db.Create() on a model that
// embeds BaseModel. If the caller didn't set an ID expclictly, we generate a v7 UUID.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		b.ID = id.String()
	}
	return nil
}

// Timestamps is embedded by every model that tracks creation + last edit.
type Timestamps struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// SoftDelete is embedded by every model whose rows can be "deleted" by users
// without losing the data. Repositories MUST filter `WHERE deleted_at IS NULL`
// DeletedBy is a UUID FK to users; nullable because the row is alive by default.
type SoftDelete struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	DeletedBy *string    `gorm:"type:uuid" json:"deleted_by,omitempty"`
}

// IsDeleted reports whether the row is currently soft-deleted.
func (s SoftDelete) IsDeleted() bool {
	return s.DeletedAt != nil
}
