package models

import "time"

type TorrentStatus string

const (
	StatusPending     TorrentStatus = "pending"
	StatusDownloading TorrentStatus = "downloading"
	StatusReady       TorrentStatus = "ready"
	StatusError       TorrentStatus = "error"
)

type TorrentRecord struct {
	ID         uint          `gorm:"primaryKey;column:id"`
	MovieID    int           `gorm:"column:movie_id"`
	MagnetURI  string        `gorm:"column:magnet_uri;not null"`
	InfoHash   string        `gorm:"column:info_hash;uniqueIndex;not null"`
	Status     TorrentStatus `gorm:"column:status;type:torrent_status_enum;default:pending"`
	FilePath   string        `gorm:"column:file_path"`
	FileSize   int64         `gorm:"column:file_size"`
	Downloaded int64         `gorm:"column:downloaded"`
	Progress   float64       `gorm:"column:progress"`
	Source     string        `gorm:"column:source"`
	ErrorMsg   string        `gorm:"column:error_msg"`
	CreatedAt  time.Time     `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time     `gorm:"column:updated_at;autoUpdateTime"`
}

func (TorrentRecord) TableName() string { return "torrents" }
