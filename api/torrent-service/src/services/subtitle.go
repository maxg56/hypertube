package services

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"torrent-service/src/conf"
)

const openSubsBaseURL = "https://api.opensubtitles.com/api/v1"

// osClient forces HTTP/1.1 — OpenSubtitles /download returns 503 for HTTP/2 requests.
var osClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	},
}

var (
	osAPIKey   string
	osUsername string
	osPassword string

	osTokenMu  sync.RWMutex
	osToken    string
	osTokenExp time.Time
)

func init() {
	osAPIKey = os.Getenv("OPENSUBTITLES_API_KEY")
	osUsername = os.Getenv("OPENSUBTITLES_USERNAME")
	osPassword = os.Getenv("OPENSUBTITLES_PASSWORD")
}

func subtitleCacheDir() string {
	if p := os.Getenv("SUBTITLE_CACHE_PATH"); p != "" {
		return p
	}
	return "/data/subtitles"
}

func subtitleCachePath(movieID int, lang string) string {
	return filepath.Join(subtitleCacheDir(), fmt.Sprintf("%d", movieID), lang+".vtt")
}

// imdbIDForTmdb looks up the IMDb ID for a TMDB movie ID from the local DB.
func imdbIDForTmdb(tmdbID int) string {
	type row struct {
		ImdbID string `gorm:"column:imdb_id"`
	}
	var r row
	conf.DB.Raw("SELECT imdb_id FROM movies WHERE tmdb_id = ?", tmdbID).Scan(&r)
	return r.ImdbID
}

// FetchSubtitle returns the path to a cached VTT subtitle file. It downloads and
// converts the subtitle from OpenSubtitles if not already cached.
func FetchSubtitle(movieID int, lang string) (string, error) {
	if osAPIKey == "" {
		return "", fmt.Errorf("OPENSUBTITLES_API_KEY not set")
	}

	path := subtitleCachePath(movieID, lang)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if osUsername != "" && osPassword != "" {
		if err := osEnsureToken(); err != nil {
			log.Printf("subtitle: opensubtitles login failed: %v", err)
		}
	}

	fileID, err := osSearch(movieID, lang)
	if err != nil {
		// Fallback: search by IMDb ID when TMDB search returns no results.
		if imdbID := imdbIDForTmdb(movieID); imdbID != "" {
			fileID, err = osSearchByIMDb(imdbID, lang)
		}
		if err != nil {
			return "", fmt.Errorf("subtitle search: %w", err)
		}
	}

	link, err := osDownloadLink(fileID)
	if err != nil {
		return "", fmt.Errorf("subtitle download link: %w", err)
	}

	raw, err := fetchURL(link)
	if err != nil {
		return "", fmt.Errorf("download subtitle file: %w", err)
	}

	vtt := srtToVTT(raw)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create subtitle cache dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(vtt), 0644); err != nil {
		return "", fmt.Errorf("write subtitle cache: %w", err)
	}

	return path, nil
}

func osEnsureToken() error {
	osTokenMu.RLock()
	valid := osToken != "" && time.Now().Before(osTokenExp)
	osTokenMu.RUnlock()
	if valid {
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"username": osUsername,
		"password": osPassword,
	})
	req, _ := http.NewRequest(http.MethodPost, openSubsBaseURL+"/login", bytes.NewReader(body))
	req.Header.Set("Api-Key", osAPIKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; hypertube/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := osClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Token == "" {
		return fmt.Errorf("empty token from opensubtitles login")
	}

	osTokenMu.Lock()
	osToken = result.Token
	osTokenExp = time.Now().Add(23 * time.Hour)
	osTokenMu.Unlock()
	return nil
}

type osSubtitleSearchResponse struct {
	Data []struct {
		Attributes struct {
			Files []struct {
				FileID int `json:"file_id"`
			} `json:"files"`
		} `json:"attributes"`
	} `json:"data"`
}

func osSearchURL(query, lang string) (int, error) {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/subtitles?%s&languages=%s&type=movie", openSubsBaseURL, query, lang), nil)
	osSetHeaders(req)

	resp, err := osClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result osSubtitleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if len(result.Data) == 0 || len(result.Data[0].Attributes.Files) == 0 {
		return 0, fmt.Errorf("no subtitles found for language %q", lang)
	}
	return result.Data[0].Attributes.Files[0].FileID, nil
}

func osSearch(movieID int, lang string) (int, error) {
	return osSearchURL(fmt.Sprintf("tmdb_id=%d", movieID), lang)
}

func osSearchByIMDb(imdbID, lang string) (int, error) {
	return osSearchURL(fmt.Sprintf("imdb_id=%s", imdbID), lang)
}

func osDownloadLink(fileID int) (string, error) {
	body, _ := json.Marshal(map[string]int{"file_id": fileID})
	req, _ := http.NewRequest(http.MethodPost, openSubsBaseURL+"/download", bytes.NewReader(body))
	osSetHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := osClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Link == "" {
		return "", fmt.Errorf("empty download link from opensubtitles (status %d)", resp.StatusCode)
	}
	return result.Link, nil
}

func osSetHeaders(req *http.Request) {
	req.Header.Set("Api-Key", osAPIKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; hypertube/1.0)")
	req.Header.Set("Accept", "application/json")
	osTokenMu.RLock()
	tok := osToken
	osTokenMu.RUnlock()
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
}

func fetchURL(url string) ([]byte, error) {
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// srtToVTT converts SRT subtitle bytes to WebVTT format.
func srtToVTT(srt []byte) string {
	var out strings.Builder
	out.WriteString("WEBVTT\n\n")
	scanner := bufio.NewScanner(bytes.NewReader(srt))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, " --> ") {
			line = strings.ReplaceAll(line, ",", ".")
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.String()
}
