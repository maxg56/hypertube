package main

import (
	"encoding/json"
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


type deleteTorrentRequest struct {
	TorrentID int    `json:"torrent_id"`
	InfoHash  string `json:"info_hash"`
	FilePath  string `json:"file_path"`
}

func downloadDir() string {
	if d := os.Getenv("TORRENT_DOWNLOAD_PATH"); d != "" {
		return d
	}
	return "/data/torrents"
}

func handleDeleteTorrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deleteTorrentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TorrentID == 0 || req.InfoHash == "" {
		http.Error(w, "torrent_id and info_hash are required", http.StatusBadRequest)
		return
	}

	torrentDir := downloadDir() + "/" + req.InfoHash
	if err := os.RemoveAll(torrentDir); err != nil && !os.IsNotExist(err) {
		log.Printf("[worker] failed to remove torrent dir %s: %v", torrentDir, err)
	}
	if req.FilePath != "" {
		if err := os.RemoveAll(req.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("[worker] failed to remove file %s: %v", req.FilePath, err)
		}
	}

	if err := markTorrentAsExpired(req.TorrentID); err != nil {
		log.Printf("[worker] db update failed for torrent_id=%d: %v", req.TorrentID, err)
		http.Error(w, "db update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[worker] deleted torrent_id=%d info_hash=%s", req.TorrentID, req.InfoHash)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

func main() {
	conf.InitDB()

	c := cron.New()
	_, err := c.AddFunc("*/10 * * * *", cleanupUnwatchedMovies)
	if err != nil {
		log.Fatalf("[MAIN] erreur configuration cron: %v", err)
	}
	c.Start()
	defer c.Stop()
	log.Println("cron démarré")

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
	mux.HandleFunc("/delete-torrent", handleDeleteTorrent)

	log.Printf("worker-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
