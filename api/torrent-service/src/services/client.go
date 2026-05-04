package services

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"

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
		downloadPath := downloadDir()

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

func downloadDir() string {
	if d := os.Getenv("TORRENT_DOWNLOAD_PATH"); d != "" {
		return d
	}
	return "/data/torrents"
}
