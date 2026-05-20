package services

import (
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"worker-service/src/conf"
	"worker-service/src/models"
	"worker-service/src/utils"
)

func expiryDuration() time.Duration {
	if v := os.Getenv("CLEANUP_EXPIRY_MINUTES"); v != "" {
		if mins, err := strconv.Atoi(v); err == nil && mins > 0 {
			return time.Duration(mins) * time.Minute
		}
	}
	if v := os.Getenv("CLEANUP_EXPIRY_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return 30 * 24 * time.Hour
}

type expiredTorrent struct {
	TorrentID int
	MovieID   int
	InfoHash  string
	FilePath  string
	UserID    int
	Title     string
}

func fetchExpiredTorrents() ([]expiredTorrent, error) {
	expiry := time.Now().Add(-expiryDuration())

	var results []expiredTorrent
	err := conf.DB.
		Model(&models.Torrent{}).
		Select("torrents.id AS torrent_id, torrents.movie_id, torrents.info_hash, torrents.file_path, MAX(watch_history.user_id) AS user_id, movies.title").
		Joins("JOIN movies ON movies.id = torrents.movie_id").
		Joins("JOIN watch_history ON watch_history.movie_id = torrents.movie_id").
		Where("torrents.status = ? AND torrents.file_path IS NOT NULL AND torrents.file_path != ''", models.StatusReady).
		Group("torrents.id, torrents.movie_id, torrents.file_path, movies.title").
		Having("MAX(watch_history.watched_at) < ?", expiry).
		Scan(&results).Error

	return results, err
}

func MarkTorrentAsDeleted(torrentID int) error {
	return conf.DB.
		Model(&models.Torrent{}).
		Where("id = ?", torrentID).
		Updates(map[string]any{
			"status":     gorm.Expr("?::torrent_status_enum", models.StatusDeleted),
			"file_path":  nil,
			"downloaded": 0,
			"progress":   0,
			"updated_at": time.Now(),
		}).Error
}

func CleanupUnwatchedMovies() {
	log.Println("[CRON] démarrage nettoyage des films expirés")

	torrents, err := fetchExpiredTorrents()
	if err != nil {
		log.Printf("[CRON] erreur récupération films expirés: %v", err)
		return
	}
	if len(torrents) == 0 {
		log.Println("[CRON] aucun film expiré")
		return
	}

	log.Printf("[CRON] %d film(s) expiré(s) trouvé(s)", len(torrents))

	var success, failed int
	for _, t := range torrents {
		utils.DeleteFromDisk(
			utils.DownloadDir()+"/"+t.InfoHash,
			t.FilePath,
		)

		if err := MarkTorrentAsDeleted(t.TorrentID); err != nil {
			log.Printf("[CRON] erreur BDD — torrent_id=%d | movie_id=%d | err=%v",
				t.TorrentID, t.MovieID, err)
			failed++
			continue
		}

		log.Printf("[CRON] supprimé — torrent_id=%d | movie_id=%d | user_id=%d | path=%s",
			t.TorrentID, t.MovieID, t.UserID, t.FilePath)
		success++
	}

	log.Printf("[CRON] terminé — succès: %d | échecs: %d", success, failed)
}
