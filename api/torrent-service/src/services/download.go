package services

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"gorm.io/gorm"

	"torrent-service/src/conf"
	"torrent-service/src/models"
)

// StartDownload adds a magnet URI to the client and begins downloading.
// It is idempotent: calling it twice with the same magnet returns the existing state.
func StartDownload(magnetURI string, movieID int) (string, error) {
	// Validate the magnet URI first so callers get a clear 4xx error for bad input
	// regardless of whether the client is ready.
	infoHash, err := extractInfoHash(magnetURI)
	if err != nil {
		return "", fmt.Errorf("invalid magnet URI: %w", err)
	}

	if client == nil {
		return "", errors.New("torrent client not initialized")
	}

	if _, ok := activeTorrents.Load(infoHash); ok {
		return infoHash, nil
	}

	record, err := findOrCreateRecord(magnetURI, infoHash, movieID)
	if err != nil {
		return "", err
	}

	t, err := addToClient(magnetURI)
	if err != nil {
		conf.DB.Model(record).Updates(map[string]any{
			"status":    models.StatusError,
			"error_msg": err.Error(),
		})
		return "", err
	}

	activeTorrents.Store(infoHash, t)
	go monitorTorrent(t, record)
	return infoHash, nil
}

func findOrCreateRecord(magnetURI, infoHash string, movieID int) (*models.TorrentRecord, error) {
	var record models.TorrentRecord
	dbErr := conf.DB.Where("info_hash = ?", infoHash).First(&record).Error

	if dbErr == nil {
		if record.Status == models.StatusReady {
			if _, statErr := os.Stat(record.FilePath); statErr == nil {
				return &record, nil
			}
		} else if record.Status == models.StatusDownloading || record.Status == models.StatusPending {
			return &record, nil
		}
		// Status is error or file missing — reset and retry.
		conf.DB.Model(&record).Updates(map[string]any{"status": models.StatusPending, "error_msg": ""})
		return &record, nil
	}

	if !errors.Is(dbErr, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("db lookup: %w", dbErr)
	}

	record = models.TorrentRecord{
		MovieID:   movieID,
		MagnetURI: magnetURI,
		InfoHash:  infoHash,
		Status:    models.StatusPending,
	}
	if err := conf.DB.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("db insert: %w", err)
	}
	return &record, nil
}

func extractInfoHash(magnetURI string) (string, error) {
	if !strings.HasPrefix(magnetURI, "magnet:") {
		return "", errors.New("only magnet URIs are supported")
	}
	spec, err := torrent.TorrentSpecFromMagnetUri(magnetURI)
	if err != nil {
		return "", err
	}
	return strings.ToLower(spec.InfoHash.HexString()), nil
}

func addToClient(uri string) (*torrent.Torrent, error) {
	if strings.HasPrefix(uri, "magnet:") {
		return client.AddMagnet(uri)
	}
	return addTorrentFromURL(uri)
}

func validateTorrentURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http/https URLs are allowed")
	}
	host := parsed.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("could not resolve host: %w", err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return errors.New("URL resolves to a private or reserved address")
		}
	}
	return nil
}

var torrentHTTPClient = &http.Client{Timeout: 30 * time.Second}

func addTorrentFromURL(rawURL string) (*torrent.Torrent, error) {
	if err := validateTorrentURL(rawURL); err != nil {
		return nil, fmt.Errorf("rejected torrent URL: %w", err)
	}
	resp, err := torrentHTTPClient.Get(rawURL)
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
