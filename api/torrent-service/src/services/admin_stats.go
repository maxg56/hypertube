package services

import (
	"fmt"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

type AdminStats struct {
	TotalUsers            int64        `json:"total_users"`
	TotalFilms            int64        `json:"total_films"`
	TotalWatches          int64        `json:"total_watches"`
	FilmsByStatus         []StatusCount `json:"films_by_status"`
	WatchesPerDay         []DayCount    `json:"watches_per_day"`
	TopFilms              []TopFilm     `json:"top_films"`
	RegistrationsPerMonth []MonthCount  `json:"registrations_per_month"`
}

type StatusCount struct {
	Status string `json:"status" gorm:"column:status"`
	Count  int64  `json:"count"  gorm:"column:count"`
}

type DayCount struct {
	Day   string `json:"day"   gorm:"column:day"`
	Count int64  `json:"count" gorm:"column:count"`
}

type MonthCount struct {
	Month string `json:"month" gorm:"column:month"`
	Count int64  `json:"count" gorm:"column:count"`
}

type TopFilm struct {
	Title   string `json:"title"   gorm:"column:title"`
	Watches int64  `json:"watches" gorm:"column:watches"`
}

func GetAdminStats() (AdminStats, error) {
	var stats AdminStats

	if err := conf.DB.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return stats, fmt.Errorf("count users: %w", err)
	}

	if err := conf.DB.Model(&models.TorrentRecord{}).Count(&stats.TotalFilms).Error; err != nil {
		return stats, fmt.Errorf("count films: %w", err)
	}

	if err := conf.DB.Model(&models.WatchHistory{}).Count(&stats.TotalWatches).Error; err != nil {
		return stats, fmt.Errorf("count watches: %w", err)
	}

	if err := conf.DB.Model(&models.TorrentRecord{}).
		Select("status::text AS status, COUNT(*) AS count").
		Group("status").
		Scan(&stats.FilmsByStatus).Error; err != nil {
		return stats, fmt.Errorf("films by status: %w", err)
	}

	if err := conf.DB.Model(&models.WatchHistory{}).
		Select("to_char(watched_at, 'YYYY-MM-DD') AS day, COUNT(*) AS count").
		Where("watched_at >= NOW() - INTERVAL '30 days'").
		Group("to_char(watched_at, 'YYYY-MM-DD')").
		Order("day ASC").
		Scan(&stats.WatchesPerDay).Error; err != nil {
		return stats, fmt.Errorf("watches per day: %w", err)
	}

	if err := conf.DB.Model(&models.WatchHistory{}).
		Select("COALESCE(m.title, t.info_hash) AS title, COUNT(watch_history.id) AS watches").
		Joins("LEFT JOIN torrents t ON t.movie_id = watch_history.movie_id").
		Joins("LEFT JOIN movies m ON m.id = watch_history.movie_id").
		Group("COALESCE(m.title, t.info_hash)").
		Order("watches DESC").
		Limit(5).
		Scan(&stats.TopFilms).Error; err != nil {
		return stats, fmt.Errorf("top films: %w", err)
	}

	if err := conf.DB.Model(&models.User{}).
		Select("to_char(created_at, 'YYYY-MM') AS month, COUNT(*) AS count").
		Where("created_at >= NOW() - INTERVAL '6 months'").
		Group("to_char(created_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&stats.RegistrationsPerMonth).Error; err != nil {
		return stats, fmt.Errorf("registrations per month: %w", err)
	}

	return stats, nil
}
