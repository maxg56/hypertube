package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/anacrolix/torrent"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

func monitorTorrent(t *torrent.Torrent, record *models.TorrentRecord) {
	infoHash := record.InfoHash

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
	filePath := resolveFilePath(t)

	conf.DB.Model(record).Updates(map[string]any{
		"status":    models.StatusDownloading,
		"file_size": totalLength,
		"file_path": filePath,
	})

	if err := runProgressLoop(t, record, totalLength, infoHash); err != nil {
		return
	}

	conf.DB.Model(record).Updates(map[string]any{
		"status":     models.StatusReady,
		"progress":   100.0,
		"downloaded": totalLength,
		"file_path":  filePath,
	})
	log.Printf("torrent %s download complete: %s", infoHash, filePath)
}

func runProgressLoop(t *torrent.Torrent, record *models.TorrentRecord, totalLength int64, infoHash string) error {
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

		if staleFor >= 120 { // 10 minutes stalled
			setError(record, "download stalled: no progress for 10 minutes")
			activeTorrents.Delete(infoHash)
			return errStalled
		}

		if totalLength > 0 && downloaded >= totalLength {
			break
		}
	}
	return nil
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

	const eagerBytes = 5 << 20
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

func resolveFilePath(t *torrent.Torrent) string {
	f := largestFile(t)
	if f == nil {
		return ""
	}
	return downloadDir() + "/" + strings.ToLower(t.InfoHash().HexString()) + "/" + f.DisplayPath()
}

func setError(record *models.TorrentRecord, msg string) {
	conf.DB.Model(record).Updates(map[string]any{
		"status":    models.StatusError,
		"error_msg": msg,
	})
	log.Printf("torrent %s error: %s", record.InfoHash, msg)
}

// sentinel so runProgressLoop can signal early exit to monitorTorrent
var errStalled = errors.New("stalled")
