package models

import "time"

type WatchHistory struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	UserID      int       `gorm:"column:user_id;not null"`
	MovieID     int       `gorm:"column:movie_id;not null"`
	WatchedAt   time.Time `gorm:"column:watched_at;autoUpdateTime"`
	ProgressSec int       `gorm:"column:progress_sec;default:0"`
}

func (WatchHistory) TableName() string { return "watch_history" }
