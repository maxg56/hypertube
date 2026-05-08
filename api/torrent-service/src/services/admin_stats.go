package services

import (
	"fmt"

	"torrent-service/src/conf"
)

type AdminStats struct {
	TotalUsers      int64            `json:"total_users"`
	TotalFilms      int64            `json:"total_films"`
	TotalWatches    int64            `json:"total_watches"`
	FilmsByStatus   []StatusCount    `json:"films_by_status"`
	WatchesPerDay   []DayCount       `json:"watches_per_day"`
	TopFilms        []TopFilm        `json:"top_films"`
	RegistrationsPerMonth []MonthCount `json:"registrations_per_month"`
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
	Title  string `json:"title"  gorm:"column:title"`
	Watches int64 `json:"watches" gorm:"column:watches"`
}

func GetAdminStats() (AdminStats, error) {
	var stats AdminStats

	if err := conf.DB.Raw("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers).Error; err != nil {
		return stats, fmt.Errorf("count users: %w", err)
	}

	if err := conf.DB.Raw("SELECT COUNT(*) FROM torrents").Scan(&stats.TotalFilms).Error; err != nil {
		return stats, fmt.Errorf("count films: %w", err)
	}

	if err := conf.DB.Raw("SELECT COUNT(*) FROM watch_history").Scan(&stats.TotalWatches).Error; err != nil {
		return stats, fmt.Errorf("count watches: %w", err)
	}

	if err := conf.DB.Raw(`
		SELECT status, COUNT(*) AS count
		FROM torrents
		GROUP BY status
	`).Scan(&stats.FilmsByStatus).Error; err != nil {
		return stats, fmt.Errorf("films by status: %w", err)
	}

	if err := conf.DB.Raw(`
		SELECT to_char(watched_at, 'YYYY-MM-DD') AS day, COUNT(*) AS count
		FROM watch_history
		WHERE watched_at >= NOW() - INTERVAL '30 days'
		GROUP BY to_char(watched_at, 'YYYY-MM-DD')
		ORDER BY day ASC
	`).Scan(&stats.WatchesPerDay).Error; err != nil {
		return stats, fmt.Errorf("watches per day: %w", err)
	}

	if err := conf.DB.Raw(`
		SELECT COALESCE(m.title, t.info_hash) AS title, COUNT(wh.id) AS watches
		FROM watch_history wh
		LEFT JOIN torrents t ON t.movie_id = wh.movie_id
		LEFT JOIN movies m ON m.id = wh.movie_id
		GROUP BY COALESCE(m.title, t.info_hash)
		ORDER BY watches DESC
		LIMIT 5
	`).Scan(&stats.TopFilms).Error; err != nil {
		return stats, fmt.Errorf("top films: %w", err)
	}

	if err := conf.DB.Raw(`
		SELECT to_char(created_at, 'YYYY-MM') AS month, COUNT(*) AS count
		FROM users
		WHERE created_at >= NOW() - INTERVAL '6 months'
		GROUP BY to_char(created_at, 'YYYY-MM')
		ORDER BY month ASC
	`).Scan(&stats.RegistrationsPerMonth).Error; err != nil {
		return stats, fmt.Errorf("registrations per month: %w", err)
	}

	return stats, nil
}
