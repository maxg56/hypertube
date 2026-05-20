package services

import (
	"log"

	"worker-service/src/conf"
	"worker-service/src/utils"
)

type DownloadedMovie struct {
	TorrentID int
	MovieID   int
	FilePath  string
	UserID    int
	Title     string
}

func fetchExpiredMovies() ([]DownloadedMovie, error) {
	query := `
		SELECT
			t.id                    AS torrent_id,
			t.movie_id,
			t.file_path,
			MAX(wh.user_id)         AS user_id,
			m.title,
			MAX(wh.watched_at)      AS last_watched_at
		FROM torrents t
		JOIN movies m ON m.id = t.movie_id
		JOIN watch_history wh ON wh.movie_id = t.movie_id
		WHERE t.file_path IS NOT NULL
		  AND t.status = 'ready'
		  AND t.file_path != ''
		GROUP BY t.id, t.movie_id, t.file_path, m.title
		HAVING MAX(wh.watched_at) < NOW() - INTERVAL '2 minutes'
	`
	var movies []DownloadedMovie
	return movies, conf.DB.Raw(query).Scan(&movies).Error
}

func MarkTorrentAsDeleted(torrentID int) error {
	return conf.DB.Exec(`
		UPDATE torrents
		SET file_path  = NULL,
		    status     = 'suppr',
		    downloaded = 0,
		    progress   = 0,
		    updated_at = NOW()
		WHERE id = ?
	`, torrentID).Error
}

func CleanupUnwatchedMovies() {
	log.Println("[CRON] démarrage nettoyage des films expirés")

	movies, err := fetchExpiredMovies()
	if err != nil {
		log.Printf("[CRON] erreur récupération films expirés: %v", err)
		return
	}
	if len(movies) == 0 {
		log.Println("[CRON] aucun film expiré")
		return
	}

	log.Printf("[CRON] %d film(s) expiré(s) trouvé(s)", len(movies))

	var success, failed int
	for _, movie := range movies {
		utils.DeleteFromDisk(movie.FilePath)

		if err := MarkTorrentAsDeleted(movie.TorrentID); err != nil {
			log.Printf("[CRON] erreur BDD — torrent_id=%d | movie_id=%d | err=%v",
				movie.TorrentID, movie.MovieID, err)
			failed++
			continue
		}

		log.Printf("[CRON] supprimé — torrent_id=%d | movie_id=%d | user_id=%d | path=%s",
			movie.TorrentID, movie.MovieID, movie.UserID, movie.FilePath)
		success++
	}

	log.Printf("[CRON] terminé — succès: %d | échecs: %d", success, failed)
}
