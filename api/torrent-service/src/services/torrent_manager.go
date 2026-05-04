package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"gorm.io/gorm"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

var (
	client         *torrent.Client
	clientOnce     sync.Once
	activeTorrents sync.Map // info_hash (lowercase) -> *torrent.Torrent
)

func InitTorrentClient() error {
	var initErr error
	clientOnce.Do(func() {
		downloadPath := os.Getenv("TORRENT_DOWNLOAD_PATH")
		if downloadPath == "" {
			downloadPath = "/data/torrents"
		}

		cfg := torrent.NewDefaultClientConfig()
		cfg.DataDir = downloadPath
		cfg.DefaultStorage = storage.NewFileByInfoHash(downloadPath)
		cfg.Seed = false

		c, err := torrent.NewClient(cfg)
		if err != nil {
			initErr = fmt.Errorf("torrent client init: %w", err)
			return
		}
		client = c
		log.Println("Torrent client initialized, data dir:", downloadPath)

		go reattachPendingTorrents()
	})
	return initErr
}

// StartDownload adds a magnet/torrent URL to the client and begins downloading.
// It is idempotent: calling it twice with the same magnet returns the existing state.
func StartDownload(magnetURI string, movieID int) (string, error) {
	if client == nil {
		return "", errors.New("torrent client not initialized")
	}

	// Derive the info hash from the magnet URI so we can check for duplicates
	// before hitting the network.
	infoHash, err := extractInfoHash(magnetURI)
	if err != nil {
		return "", fmt.Errorf("invalid magnet URI: %w", err)
	}

	// Already active in this process — just return.
	if _, ok := activeTorrents.Load(infoHash); ok {
		return infoHash, nil
	}

	// Check DB for an existing record.
	var record models.TorrentRecord
	dbErr := conf.DB.Where("info_hash = ?", infoHash).First(&record).Error
	if dbErr == nil {
		// Already in DB.
		if record.Status == models.StatusReady {
			if _, statErr := os.Stat(record.FilePath); statErr == nil {
				return infoHash, nil
			}
			// File missing — fall through to re-download.
		} else if record.Status == models.StatusDownloading || record.Status == models.StatusPending {
			// Reattach if not active (e.g. after restart when reattach is still in flight).
			return infoHash, nil
		}
		// Status is error or file missing — reset and retry.
		conf.DB.Model(&record).Updates(map[string]any{
			"status":    models.StatusPending,
			"error_msg": "",
		})
	} else if !errors.Is(dbErr, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("db lookup: %w", dbErr)
	} else {
		// Create new record.
		record = models.TorrentRecord{
			MovieID:   movieID,
			MagnetURI: magnetURI,
			InfoHash:  infoHash,
			Status:    models.StatusPending,
		}
		if createErr := conf.DB.Create(&record).Error; createErr != nil {
			return "", fmt.Errorf("db insert: %w", createErr)
		}
	}

	t, err := addToClient(magnetURI)
	if err != nil {
		conf.DB.Model(&record).Updates(map[string]any{
			"status":    models.StatusError,
			"error_msg": err.Error(),
		})
		return "", err
	}

	activeTorrents.Store(infoHash, t)
	go monitorTorrent(t, &record)
	return infoHash, nil
}

// GetTorrentReader returns an io.ReadSeeker for the largest file in the torrent.
// For a completed torrent it reads from disk; for an in-progress one it uses
// the anacrolix reader, which blocks on missing pieces naturally.
func GetTorrentReader(infoHash string) (io.ReadSeeker, int64, string, error) {
	// Try in-memory active torrent first.
	if v, ok := activeTorrents.Load(infoHash); ok {
		t := v.(*torrent.Torrent)
		select {
		case <-t.GotInfo():
		case <-time.After(30 * time.Second):
			return nil, 0, "", errors.New("torrent info not yet available")
		}
		f := largestFile(t)
		if f == nil {
			return nil, 0, "", errors.New("no files in torrent")
		}
		r := f.NewReader()
		r.SetReadahead(5 << 20) // 5 MiB readahead for smooth streaming
		return r, f.Length(), f.DisplayPath(), nil
	}

	// Fall back to disk (completed torrent, process restarted).
	var record models.TorrentRecord
	if err := conf.DB.Where("info_hash = ?", infoHash).First(&record).Error; err != nil {
		return nil, 0, "", fmt.Errorf("torrent not found: %w", err)
	}
	if record.Status != models.StatusReady || record.FilePath == "" {
		return nil, 0, "", errors.New("torrent not ready")
	}
	f, err := os.Open(record.FilePath)
	if err != nil {
		return nil, 0, "", fmt.Errorf("open file: %w", err)
	}
	return f, record.FileSize, record.FilePath, nil
}

// GetRecord fetches the DB record for a given info hash.
func GetRecord(infoHash string) (*models.TorrentRecord, error) {
	var record models.TorrentRecord
	if err := conf.DB.Where("info_hash = ?", infoHash).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ---- internal helpers ----

func extractInfoHash(magnetURI string) (string, error) {
	// Support magnet links and plain .torrent HTTP URLs.
	if strings.HasPrefix(magnetURI, "magnet:") {
		spec, err := torrent.TorrentSpecFromMagnetUri(magnetURI)
		if err != nil {
			return "", err
		}
		return strings.ToLower(spec.InfoHash.HexString()), nil
	}
	// For .torrent URLs we need to download to get the hash; use a placeholder
	// that will be resolved after addToClient.
	return "", errors.New("only magnet URIs are supported")
}

func addToClient(uri string) (*torrent.Torrent, error) {
	if strings.HasPrefix(uri, "magnet:") {
		t, err := client.AddMagnet(uri)
		return t, err
	}
	// Download .torrent file and add it.
	resp, err := http.Get(uri) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("fetch torrent file: %w", err)
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp("", "*.torrent")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err = io.Copy(tmp, resp.Body); err != nil {
		return nil, err
	}
	tmp.Close()
	return client.AddTorrentFromFile(tmp.Name())
}

func monitorTorrent(t *torrent.Torrent, record *models.TorrentRecord) {
	infoHash := record.InfoHash

	// Wait for metadata with a 2-minute timeout.
	select {
	case <-t.GotInfo():
	case <-time.After(2 * time.Minute):
		setError(record, "no seeders or peers: info timeout")
		activeTorrents.Delete(infoHash)
		return
	}

	t.DownloadAll()
	prioritizeForStreaming(t)

	totalLength := t.Info().TotalLength()
	videoFile := largestFile(t)
	filePath := ""
	if videoFile != nil {
		downloadPath := os.Getenv("TORRENT_DOWNLOAD_PATH")
		if downloadPath == "" {
			downloadPath = "/data/torrents"
		}
		filePath = downloadPath + "/" + strings.ToLower(t.InfoHash().HexString()) + "/" + videoFile.DisplayPath()
	}

	conf.DB.Model(record).Updates(map[string]any{
		"status":    models.StatusDownloading,
		"file_size": totalLength,
		"file_path": filePath,
	})

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastDownloaded int64
	staleFor := 0

	for range ticker.C {
		downloaded := t.BytesCompleted()
		progress := 0.0
		if totalLength > 0 {
			progress = float64(downloaded) / float64(totalLength) * 100
		}

		conf.DB.Model(record).Updates(map[string]any{
			"downloaded": downloaded,
			"progress":   progress,
		})

		if downloaded == lastDownloaded {
			staleFor++
		} else {
			staleFor = 0
			lastDownloaded = downloaded
		}
		// 10 minutes with no progress → error
		if staleFor >= 120 {
			setError(record, "download stalled: no progress for 10 minutes")
			activeTorrents.Delete(infoHash)
			return
		}

		if t.BytesCompleted() >= totalLength && totalLength > 0 {
			break
		}
	}

	conf.DB.Model(record).Updates(map[string]any{
		"status":     models.StatusReady,
		"progress":   100.0,
		"downloaded": totalLength,
		"file_path":  filePath,
	})
	log.Printf("torrent %s download complete: %s", infoHash, filePath)
}

func prioritizeForStreaming(t *torrent.Torrent) {
	f := largestFile(t)
	if f == nil {
		return
	}
	f.SetPriority(torrent.PiecePriorityNormal)

	pieceLen := int64(t.Info().PieceLength)
	if pieceLen == 0 {
		return
	}

	const eagerBytes = 5 << 20 // 5 MiB
	eagerPieces := (eagerBytes / pieceLen) + 1
	pieceOffset := f.Offset() / pieceLen

	for i := int64(0); i < eagerPieces; i++ {
		idx := int(pieceOffset + i)
		if idx < t.NumPieces() {
			t.Piece(idx).SetPriority(torrent.PiecePriorityNow)
		}
	}
}

func largestFile(t *torrent.Torrent) *torrent.File {
	var best *torrent.File
	for _, f := range t.Files() {
		f := f
		if best == nil || f.Length() > best.Length() {
			best = f
		}
	}
	return best
}

func setError(record *models.TorrentRecord, msg string) {
	conf.DB.Model(record).Updates(map[string]any{
		"status":    models.StatusError,
		"error_msg": msg,
	})
	log.Printf("torrent %s error: %s", record.InfoHash, msg)
}

func reattachPendingTorrents() {
	var records []models.TorrentRecord
	conf.DB.Where("status IN ?", []string{
		string(models.StatusPending),
		string(models.StatusDownloading),
	}).Find(&records)

	for i := range records {
		r := &records[i]
		t, err := addToClient(r.MagnetURI)
		if err != nil {
			setError(r, err.Error())
			continue
		}
		activeTorrents.Store(r.InfoHash, t)
		go monitorTorrent(t, r)
		log.Printf("reattached torrent %s", r.InfoHash)
	}
}
