package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/anacrolix/torrent"
	"gorm.io/gorm"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// AdminTorrentEntry holds per-torrent info inside a grouped film.
type AdminTorrentEntry struct {
	ID         uint    `json:"id"`
	InfoHash   string  `json:"info_hash"`
	Status     string  `json:"status"`
	FileSize   int64   `json:"file_size"`
	Downloaded int64   `json:"downloaded"`
	Progress   float64 `json:"progress"`
	Quality    string  `json:"quality"`
	CreatedAt  string  `json:"created_at"`
}

// AdminGroupedFilm groups all torrents for a single movie.
type AdminGroupedFilm struct {
	MovieID       int                 `json:"movie_id"`
	TmdbID        int                 `json:"tmdb_id"`
	Title         string              `json:"title"`
	PosterPath    string              `json:"poster_path"`
	WatchersCount int64               `json:"watchers_count"`
	WatcherIDs    []int               `json:"watcher_ids"`
	Torrents      []AdminTorrentEntry `json:"torrents"`
}

// ListAdminFilms returns torrent records grouped by movie with watcher info.
// Uses 4 focused GORM queries assembled in Go instead of a raw SQL CTE.
func ListAdminFilms(limit, offset int) ([]AdminGroupedFilm, int64, error) {
	var total int64
	if err := conf.DB.Model(&models.TorrentRecord{}).Distinct("movie_id").Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}
	if total == 0 {
		return []AdminGroupedFilm{}, 0, nil
	}

	// 1. Paginated list of movie_ids ordered by most recent torrent.
	type movieGroup struct {
		MovieID    int       `gorm:"column:movie_id"`
		MaxCreated time.Time `gorm:"column:max_created"`
	}
	var groups []movieGroup
	if err := conf.DB.Model(&models.TorrentRecord{}).
		Select("movie_id, MAX(created_at) AS max_created").
		Group("movie_id").
		Order("max_created DESC").
		Limit(limit).
		Offset(offset).
		Scan(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("fetch groups: %w", err)
	}

	movieIDs := make([]int, len(groups))
	for i, g := range groups {
		movieIDs[i] = g.MovieID
	}

	// 2. Movie metadata for those IDs.
	var movies []models.Movie
	if err := conf.DB.Where("id IN ?", movieIDs).Find(&movies).Error; err != nil {
		return nil, 0, fmt.Errorf("fetch movies: %w", err)
	}
	movieByID := make(map[int]models.Movie, len(movies))
	for _, m := range movies {
		movieByID[m.ID] = m
	}

	// 3. All torrent records for those movies.
	var torrents []models.TorrentRecord
	if err := conf.DB.Where("movie_id IN ?", movieIDs).Find(&torrents).Error; err != nil {
		return nil, 0, fmt.Errorf("fetch torrents: %w", err)
	}
	torrentsByMovie := make(map[int][]models.TorrentRecord, len(groups))
	for _, t := range torrents {
		torrentsByMovie[t.MovieID] = append(torrentsByMovie[t.MovieID], t)
	}

	// 4. Unique watchers per movie (watch_history has UNIQUE(user_id, movie_id)).
	type watchRow struct {
		MovieID int `gorm:"column:movie_id"`
		UserID  int `gorm:"column:user_id"`
	}
	var watchRows []watchRow
	if err := conf.DB.Model(&models.WatchHistory{}).
		Select("movie_id, user_id").
		Where("movie_id IN ?", movieIDs).
		Scan(&watchRows).Error; err != nil {
		return nil, 0, fmt.Errorf("fetch watchers: %w", err)
	}
	watchersByMovie := make(map[int][]int, len(groups))
	for _, w := range watchRows {
		watchersByMovie[w.MovieID] = append(watchersByMovie[w.MovieID], w.UserID)
	}

	// Assemble result preserving the pagination order.
	result := make([]AdminGroupedFilm, 0, len(groups))
	for _, g := range groups {
		movie := movieByID[g.MovieID]

		recs := torrentsByMovie[g.MovieID]
		sort.Slice(recs, func(i, j int) bool {
			return recs[i].CreatedAt.After(recs[j].CreatedAt)
		})
		entries := make([]AdminTorrentEntry, 0, len(recs))
		for _, t := range recs {
			entries = append(entries, AdminTorrentEntry{
				ID:         t.ID,
				InfoHash:   t.InfoHash,
				Status:     string(t.Status),
				FileSize:   t.FileSize,
				Downloaded: t.Downloaded,
				Progress:   t.Progress,
				Quality:    t.Quality,
				CreatedAt:  t.CreatedAt.UTC().Format(time.RFC3339),
			})
		}

		watchers := watchersByMovie[g.MovieID]
		if watchers == nil {
			watchers = []int{}
		}

		result = append(result, AdminGroupedFilm{
			MovieID:       g.MovieID,
			TmdbID:        movie.TmdbID,
			Title:         movie.Title,
			PosterPath:    movie.PosterPath,
			WatchersCount: int64(len(watchers)),
			WatcherIDs:    watchers,
			Torrents:      entries,
		})
	}
	return result, total, nil
}

func workerURL() string {
	if u := os.Getenv("WORKER_SERVICE_URL"); u != "" {
		return u
	}
	return "http://worker-service:8006"
}

// DeleteAdminFilm drops the active torrent from memory then delegates file
// deletion and DB update to the worker-service, preserving watch history.
func DeleteAdminFilm(id uint) error {
	var record models.TorrentRecord
	if err := conf.DB.First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return fmt.Errorf("db lookup: %w", err)
	}

	if v, ok := activeTorrents.Load(record.InfoHash); ok {
		if t, ok := v.(*torrent.Torrent); ok {
			t.Drop()
		}
		activeTorrents.Delete(record.InfoHash)
	}

	body, _ := json.Marshal(map[string]any{
		"torrent_id": record.ID,
		"info_hash":  record.InfoHash,
		"file_path":  record.FilePath,
	})
	resp, err := http.Post(workerURL()+"/delete-torrent", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("worker request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("worker returned %d", resp.StatusCode)
	}

	log.Printf("[admin] delegated deletion of torrent %d (hash=%s) to worker", id, record.InfoHash)
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

	activeTorrents.Delete(record.InfoHash)

	conf.DB.Model(&record).Updates(map[string]any{
		"status":    models.StatusPending,
		"error_msg": "",
		"progress":  0,
	})

	var tmdbID int
	conf.DB.Model(&models.Movie{}).Where("id = ?", record.MovieID).Pluck("tmdb_id", &tmdbID)
	if tmdbID == 0 {
		tmdbID = record.MovieID
	}

	return StartDownload(record.MagnetURI, tmdbID, record.Quality)
}
