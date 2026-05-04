package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/anacrolix/torrent"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// GetTorrentReader returns an io.ReadSeeker for the largest file in the torrent.
// For an in-progress torrent it uses the anacrolix reader (blocks on missing pieces).
// For a completed torrent it reads directly from disk.
func GetTorrentReader(infoHash string) (readerResult, error) {
	if v, ok := activeTorrents.Load(infoHash); ok {
		return readerFromActiveTorrent(v.(*torrent.Torrent))
	}
	return readerFromDisk(infoHash)
}

// GetRecord fetches the DB record for a given info hash.
func GetRecord(infoHash string) (*models.TorrentRecord, error) {
	var record models.TorrentRecord
	if err := conf.DB.Where("info_hash = ?", infoHash).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

type readerResult struct {
	Reader   torrentReader
	Size     int64
	FileName string
}

// torrentReader is satisfied by both *os.File and torrent.Reader.
type torrentReader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func readerFromActiveTorrent(t *torrent.Torrent) (readerResult, error) {
	select {
	case <-t.GotInfo():
	case <-time.After(30 * time.Second):
		return readerResult{}, errors.New("torrent info not yet available")
	}

	f := largestFile(t)
	if f == nil {
		return readerResult{}, errors.New("no files in torrent")
	}

	r := f.NewReader()
	r.SetReadahead(5 << 20) // 5 MiB readahead for smooth streaming
	return readerResult{Reader: r, Size: f.Length(), FileName: f.DisplayPath()}, nil
}

func readerFromDisk(infoHash string) (readerResult, error) {
	record, err := GetRecord(infoHash)
	if err != nil {
		return readerResult{}, fmt.Errorf("torrent not found: %w", err)
	}
	if record.Status != models.StatusReady || record.FilePath == "" {
		return readerResult{}, errors.New("torrent not ready")
	}

	f, err := os.Open(record.FilePath)
	if err != nil {
		return readerResult{}, fmt.Errorf("open file: %w", err)
	}
	return readerResult{Reader: f, Size: record.FileSize, FileName: record.FilePath}, nil
}
