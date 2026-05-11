package models

import "time"

type Favorite struct {
	ID      uint      `gorm:"primaryKey"`
	UserID  int       `gorm:"column:user_id;not null"`
	TmdbID  int       `gorm:"column:tmdb_id;not null"`
	AddedAt time.Time `gorm:"column:added_at;autoCreateTime"`
}

func (Favorite) TableName() string { return "favorites" }
