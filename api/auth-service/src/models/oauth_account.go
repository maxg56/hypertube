package models

import "time"

type OAuthAccount struct {
	ID         uint      `gorm:"primaryKey;column:id"`
	UserID     uint      `gorm:"column:user_id;not null;index;type:int"`
	Provider   string    `gorm:"column:provider;type:varchar(50);not null"`
	ProviderID string    `gorm:"column:provider_id;type:varchar(255);not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (OAuthAccount) TableName() string { return "oauth_accounts" }
