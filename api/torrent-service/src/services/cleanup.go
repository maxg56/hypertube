package services

import (
	"log"
	"os"
	"time"

	"github.com/anacrolix/torrent"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

const cleanupInterval = 1 * time.Hour

func StartCleanupWorker() {
	go runCleanupLoop()
}

func runCleanupLoop() {
	purgeErrorTorrents()
	purgeRedundantQualities()
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		purgeErrorTorrents()
		purgeRedundantQualities()
	}
}

// purgeErrorTorrents removes partial files left on disk by torrents in error state.
func purgeErrorTorrents() {
	var records []models.TorrentRecord
	conf.DB.Where("status = ? AND file_path != ''", models.StatusError).Find(&records)

	for i := range records {
		r := &records[i]
		torrentDir := downloadDir() + "/" + r.InfoHash
		if err := os.RemoveAll(torrentDir); err != nil && !os.IsNotExist(err) {
			log.Printf("[cleanup] failed to remove dir %s: %v", torrentDir, err)
			continue
		}
		conf.DB.Model(r).Update("file_path", "")
		log.Printf("[cleanup] purged partial files for error torrent %s", r.InfoHash)
	}
}

// purgeRedundantQualities keeps only the smallest ready file for each movie and
// deletes the rest from disk and DB. Called after re-encoding and on the hourly tick.
func purgeRedundantQualities() {
	type movieCount struct {
		MovieID int `gorm:"column:movie_id"`
	}
	var dupes []movieCount
	conf.DB.Model(&models.TorrentRecord{}).
		Select("movie_id").
		Where("status = ? AND file_path != ''", models.StatusReady).
		Group("movie_id").
		Having("COUNT(*) > 1").
		Scan(&dupes)

	for _, d := range dupes {
		var records []models.TorrentRecord
		conf.DB.Where("movie_id = ? AND status = ? AND file_path != ''", d.MovieID, models.StatusReady).
			Order("file_size ASC").
			Find(&records)

		if len(records) <= 1 {
			continue
		}

		// Keep the first (smallest); purge the rest.
		for _, r := range records[1:] {
			dropTorrent(r.InfoHash)
			torrentDir := downloadDir() + "/" + r.InfoHash
			if err := os.RemoveAll(torrentDir); err != nil && !os.IsNotExist(err) {
				log.Printf("[dedup] failed to remove dir %s: %v", torrentDir, err)
				continue
			}
			conf.DB.Delete(&r)
			log.Printf("[dedup] movie %d: removed redundant torrent %s (quality=%s, size=%d)",
				d.MovieID, r.InfoHash, r.Quality, r.FileSize)
		}
	}
}

// dropTorrent removes a torrent from the active client if it is seeding/downloading.
func dropTorrent(infoHash string) {
	if v, ok := activeTorrents.Load(infoHash); ok {
		if t, ok := v.(*torrent.Torrent); ok {
			t.Drop()
		}
		activeTorrents.Delete(infoHash)
	}
}
