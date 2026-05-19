package main

import (
	"log"
	"net/http"
	"os"

	"worker-service/src/conf"
	"github.com/robfig/cron/v3"
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
	result := conf.DB.Raw(query).Scan(&movies)
	if result.Error != nil {
		return nil, result.Error
	}

	return movies, nil
}

func deleteFileFromDisk(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("[CRON] fichier absent sur le disque: %s", filePath)
		return nil
	}
	return os.Remove(filePath)
}

func markTorrentAsExpired(torrentID int) error {
	result := conf.DB.Exec(`
		UPDATE torrents
		SET file_path  = NULL,
		    status     = 'suppr',
		    downloaded = 0,
		    progress   = 0,
		    updated_at = NOW()
		WHERE id = ?
	`,
	torrentID)
	return result.Error
}

func cleanupUnwatchedMovies() {
	log.Println("Démarrage nettoyage des films expirés (Hypertube)")

	movies, err := fetchExpiredMovies()
	if err != nil {
		log.Printf("Erreur récupération films expirés: %v", err)
		return
	}

	if len(movies) == 0 {
		log.Println("Aucun film expiré à supprimer")
		return
	}

	log.Printf("%d film(s) expiré(s) trouvé(s)", len(movies))

	var success, failed int

	for _, movie := range movies {

		if err := deleteFileFromDisk(movie.FilePath); err != nil {
			log.Printf("ERREUR suppression fichier — torrent_id=%d | movie_id=%d | path=%s | err=%v",
				movie.TorrentID,
				movie.MovieID,
				movie.FilePath,
				err)
			failed++
			continue
		}

		if err := markTorrentAsExpired(movie.TorrentID); err != nil {
			log.Printf("ERREUR mise à jour BDD — torrent_id=%d | movie_id=%d | err=%v",
				movie.TorrentID,
				movie.MovieID,
				err)
			failed++
			continue
		}

		log.Printf("supprimé — torrent_id=%d | movie_id=%d | user_id=%d | path=%s",
			movie.TorrentID,
			movie.MovieID,
			movie.UserID,
			movie.FilePath)
		success++
	}

	log.Printf("terminé — succès: %d | échecs: %d", success, failed)
}


func main() {
	conf.InitDB();

	c := cron.New()
	_, err := c.AddFunc("*/10 * * * *", cleanupUnwatchedMovies)
	if err != nil {
		log.Fatalf("[MAIN] erreur configuration cron: %v", err)
	}
	c.Start()
	defer c.Stop()
	log.Println("cron démarré ")

	go cleanupUnwatchedMovies()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"worker-service"}`))
	})

	log.Printf("worker-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
