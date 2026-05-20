package models

import "time"

type WatchHistory struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	UserID      int       `gorm:"column:user_id"`
	MovieID     int       `gorm:"column:movie_id"`
	WatchedAt   time.Time `gorm:"column:watched_at"`
	ProgressSec int       `gorm:"column:progress_sec"`
}

func (WatchHistory) TableName() string { return "watch_history" }
