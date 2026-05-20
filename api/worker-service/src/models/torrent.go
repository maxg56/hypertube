package models

import "time"

type TorrentStatus string

const (
	StatusPending     TorrentStatus = "pending"
	StatusDownloading TorrentStatus = "downloading"
	StatusReady       TorrentStatus = "ready"
	StatusError       TorrentStatus = "error"
	StatusDeleted     TorrentStatus = "suppr"
)

type Torrent struct {
	ID         uint          `gorm:"primaryKey;column:id"`
	MovieID    int           `gorm:"column:movie_id"`
	MagnetURI  string        `gorm:"column:magnet_uri"`
	InfoHash   string        `gorm:"column:info_hash"`
	Status     TorrentStatus `gorm:"column:status;type:torrent_status_enum"`
	FilePath   string        `gorm:"column:file_path"`
	FileSize   int64         `gorm:"column:file_size"`
	Downloaded int64         `gorm:"column:downloaded"`
	Progress   float64       `gorm:"column:progress"`
	Quality    string        `gorm:"column:quality"`
	Source     string        `gorm:"column:source"`
	ErrorMsg   string        `gorm:"column:error_msg"`
	Reencoded  bool          `gorm:"column:reencoded"`
	CreatedAt  time.Time     `gorm:"column:created_at"`
	UpdatedAt  time.Time     `gorm:"column:updated_at"`
}

func (Torrent) TableName() string { return "torrents" }
