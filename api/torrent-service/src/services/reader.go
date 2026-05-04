package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/torrent"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// GetTorrentReader returns an io.ReadSeeker for the largest file in the torrent.
// For an in-progress torrent it uses the anacrolix reader (blocks on missing pieces).
// For a completed torrent it reads directly from disk.
func GetTorrentReader(infoHash string) (ReaderResult, error) {
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

type ReaderResult struct {
	Reader   torrentReader
	Size     int64
	FileName string
	FilePath string // absolute path on disk; empty while pieces are still downloading
}

// torrentReader is satisfied by both *os.File and torrent.Reader.
type torrentReader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func readerFromActiveTorrent(t *torrent.Torrent) (ReaderResult, error) {
	select {
	case <-t.GotInfo():
	case <-time.After(30 * time.Second):
		return ReaderResult{}, errors.New("torrent info not yet available")
	}

	f := largestFile(t)
	if f == nil {
		return ReaderResult{}, errors.New("no files in torrent")
	}

	r := f.NewReader()
	r.SetReadahead(5 << 20) // 5 MiB readahead for smooth streaming
	filePath := downloadDir() + "/" + strings.ToLower(t.InfoHash().HexString()) + "/" + f.DisplayPath()
	return ReaderResult{Reader: r, Size: f.Length(), FileName: f.DisplayPath(), FilePath: filePath}, nil
}

func readerFromDisk(infoHash string) (ReaderResult, error) {
	record, err := GetRecord(infoHash)
	if err != nil {
		return ReaderResult{}, fmt.Errorf("torrent not found: %w", err)
	}
	if record.Status != models.StatusReady || record.FilePath == "" {
		return ReaderResult{}, errors.New("torrent not ready")
	}

	f, err := os.Open(record.FilePath)
	if err != nil {
		return ReaderResult{}, fmt.Errorf("open file: %w", err)
	}
	return ReaderResult{Reader: f, Size: record.FileSize, FileName: record.FilePath, FilePath: record.FilePath}, nil
}
