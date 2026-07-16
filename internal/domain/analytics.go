package domain

import "time"

type UserActivityDaily struct {
	UserID       string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	ActivityDate time.Time `gorm:"type:date;primaryKey" json:"activity_date"`
}

func (UserActivityDaily) TableName() string { return "user_activity_daily" }
