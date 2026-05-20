package utils

import (
	"log"
	"os"
)

func DownloadDir() string {
	if d := os.Getenv("TORRENT_DOWNLOAD_PATH"); d != "" {
		return d
	}
	return "/data/torrents"
}

func DeleteFromDisk(paths ...string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
			log.Printf("[worker] failed to remove %s: %v", p, err)
		}
	}
}
