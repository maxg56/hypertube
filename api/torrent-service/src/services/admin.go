package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/anacrolix/torrent"
	"gorm.io/gorm"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// AdminFilmRow is the result of the admin films list query.
type AdminFilmRow struct {
	ID            uint    `json:"id"`
	MovieID       int     `json:"movie_id"`
	TmdbID        int     `json:"tmdb_id"`
	InfoHash      string  `json:"info_hash"`
	Status        string  `json:"status"`
	FilePath      string  `json:"file_path"`
	FileSize      int64   `json:"file_size"`
	Downloaded    int64   `json:"downloaded"`
	Progress      float64 `json:"progress"`
	Title         string  `json:"title"`
	PosterPath    string  `json:"poster_path"`
	Language      string  `json:"language"`
	CreatedAt     string  `json:"created_at"`
	WatchersCount int64   `json:"watchers_count"`
	// comma-separated user IDs — parsed by the handler
	watcherIDsRaw string
	WatcherIDs    []int `json:"watcher_ids"`
}

// ListAdminFilms returns all torrent records with movie metadata and watcher info.
func ListAdminFilms(limit, offset int) ([]AdminFilmRow, int64, error) {
	var total int64
	if err := conf.DB.Table("torrents").Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	type rawRow struct {
		ID            uint    `gorm:"column:id"`
		MovieID       int     `gorm:"column:movie_id"`
		TmdbID        int     `gorm:"column:tmdb_id"`
		InfoHash      string  `gorm:"column:info_hash"`
		Status        string  `gorm:"column:status"`
		FilePath      string  `gorm:"column:file_path"`
		FileSize      int64   `gorm:"column:file_size"`
		Downloaded    int64   `gorm:"column:downloaded"`
		Progress      float64 `gorm:"column:progress"`
		Title         string  `gorm:"column:title"`
		PosterPath    string  `gorm:"column:poster_path"`
		Language      string  `gorm:"column:language"`
		CreatedAt     string  `gorm:"column:created_at"`
		WatchersCount int64   `gorm:"column:watchers_count"`
		WatcherIDsRaw string  `gorm:"column:watcher_ids"`
	}

	var rows []rawRow
	err := conf.DB.Raw(`
		SELECT
			t.id,
			t.movie_id,
			COALESCE(m.tmdb_id, 0)         AS tmdb_id,
			t.info_hash,
			t.status,
			COALESCE(t.file_path, '')       AS file_path,
			t.file_size,
			t.downloaded,
			t.progress,
			COALESCE(m.title, '')           AS title,
			COALESCE(m.poster_path, '')     AS poster_path,
			COALESCE(m.language, '')        AS language,
			to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
			COUNT(DISTINCT wh.user_id)      AS watchers_count,
			COALESCE(string_agg(DISTINCT wh.user_id::text, ','), '') AS watcher_ids
		FROM torrents t
		LEFT JOIN movies m  ON m.id = t.movie_id
		LEFT JOIN watch_history wh ON wh.movie_id = t.movie_id
		GROUP BY t.id, m.tmdb_id, m.title, m.poster_path, m.language
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("query: %w", err)
	}

	result := make([]AdminFilmRow, 0, len(rows))
	for _, r := range rows {
		row := AdminFilmRow{
			ID:            r.ID,
			MovieID:       r.MovieID,
			TmdbID:        r.TmdbID,
			InfoHash:      r.InfoHash,
			Status:        r.Status,
			FilePath:      r.FilePath,
			FileSize:      r.FileSize,
			Downloaded:    r.Downloaded,
			Progress:      r.Progress,
			Title:         r.Title,
			PosterPath:    r.PosterPath,
			Language:      r.Language,
			CreatedAt:     r.CreatedAt,
			WatchersCount: r.WatchersCount,
			WatcherIDs:    parseIntCSV(r.WatcherIDsRaw),
		}
		result = append(result, row)
	}
	return result, total, nil
}

// DeleteAdminFilm removes a torrent from disk and clears all related DB records.
func DeleteAdminFilm(id uint) error {
	var record models.TorrentRecord
	if err := conf.DB.First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return fmt.Errorf("db lookup: %w", err)
	}

	// Remove active torrent from client if still seeding/downloading.
	if v, ok := activeTorrents.Load(record.InfoHash); ok {
		if t, ok := v.(*torrent.Torrent); ok {
			t.Drop()
		}
		activeTorrents.Delete(record.InfoHash)
	}

	// Remove files from disk (entire info-hash directory).
	torrentDir := downloadDir() + "/" + record.InfoHash
	if err := os.RemoveAll(torrentDir); err != nil && !os.IsNotExist(err) {
		log.Printf("[admin] warning: failed to remove torrent dir %s: %v", torrentDir, err)
	}

	// Also attempt to remove by stored file_path directory as a fallback.
	if record.FilePath != "" {
		if err := os.RemoveAll(record.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("[admin] warning: failed to remove file %s: %v", record.FilePath, err)
		}
	}

	// Clear watch_history for this movie.
	conf.DB.Where("movie_id = ?", record.MovieID).Delete(&models.WatchHistory{})

	// Delete the torrent record.
	if err := conf.DB.Delete(&record).Error; err != nil {
		return fmt.Errorf("db delete: %w", err)
	}

	log.Printf("[admin] deleted torrent %d (hash=%s)", id, record.InfoHash)
	return nil
}

// ReDownloadFilm resets a torrent record to pending and re-queues it for download.
func ReDownloadFilm(id uint) (string, error) {
	var record models.TorrentRecord
	if err := conf.DB.First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("not found")
		}
		return "", fmt.Errorf("db lookup: %w", err)
	}

	if record.MagnetURI == "" {
		return "", fmt.Errorf("no magnet URI stored for this record")
	}

	// Drop active torrent so StartDownload can re-add it.
	activeTorrents.Delete(record.InfoHash)

	// Reset to pending so findOrCreateRecord takes the retry path.
	conf.DB.Model(&record).Updates(map[string]any{
		"status":    models.StatusPending,
		"error_msg": "",
		"progress":  0,
	})

	// StartDownload expects the TMDB ID; look it up from movies table.
	var tmdbID int
	conf.DB.Raw("SELECT tmdb_id FROM movies WHERE id = ?", record.MovieID).Scan(&tmdbID)
	if tmdbID == 0 {
		tmdbID = record.MovieID // fallback: treat stored movie_id as tmdb_id
	}

	return StartDownload(record.MagnetURI, tmdbID)
}

// parseIntCSV splits a comma-separated string of integers into a slice.
func parseIntCSV(s string) []int {
	if s == "" {
		return []int{}
	}
	var result []int
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			if n := atoi(part); n > 0 {
				result = append(result, n)
			}
			start = i + 1
		}
	}
	return result
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
