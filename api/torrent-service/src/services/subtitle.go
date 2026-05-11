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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
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

	osLangCacheMu sync.RWMutex
	osLangCache   = map[int]osLangEntry{}
)

type osLangEntry struct {
	langs []string
	exp   time.Time
}

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

func SubtitleCacheDir() string { return subtitleCacheDir() }

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

// ListAvailableLanguages returns the language codes for which OpenSubtitles has
// at least one subtitle file for the given TMDB movie ID.
// Results are cached in memory for 1 hour to avoid repeated API calls.
func ListAvailableLanguages(movieID int) []string {
	osLangCacheMu.RLock()
	entry, ok := osLangCache[movieID]
	osLangCacheMu.RUnlock()
	if ok && time.Now().Before(entry.exp) {
		return entry.langs
	}

	if osAPIKey == "" {
		return nil
	}
	if osUsername != "" && osPassword != "" {
		if err := osEnsureToken(); err != nil {
			log.Printf("subtitle: opensubtitles login failed: %v", err)
		}
	}

	langs := osQueryAllLanguages(fmt.Sprintf("tmdb_id=%d", movieID))
	if len(langs) == 0 {
		if imdbID := imdbIDForTmdb(movieID); imdbID != "" {
			langs = osQueryAllLanguages(fmt.Sprintf("imdb_id=%s", imdbID))
		}
	}

	osLangCacheMu.Lock()
	osLangCache[movieID] = osLangEntry{langs: langs, exp: time.Now().Add(time.Hour)}
	osLangCacheMu.Unlock()
	log.Printf("subtitle: movie %d — available languages: %v", movieID, langs)
	return langs
}

// osQueryAllLanguages fetches all subtitles for a movie (no language filter) and
// returns the deduplicated list of ISO 639-1 language codes available.
func osQueryAllLanguages(query string) []string {
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/subtitles?%s&type=movie", openSubsBaseURL, query), nil)
	osSetHeaders(req)

	resp, err := osClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Attributes struct {
				Language string `json:"language"`
				Files    []struct {
					FileID int `json:"file_id"`
				} `json:"files"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var langs []string
	for _, d := range result.Data {
		if len(d.Attributes.Files) == 0 {
			continue
		}
		raw := strings.ToLower(d.Attributes.Language)
		code, ok := langMap[raw]
		if !ok {
			if len(raw) == 2 {
				code = raw
			} else {
				continue
			}
		}
		if !seen[code] {
			seen[code] = true
			langs = append(langs, code)
		}
	}
	return langs
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

var assTimingRe = regexp.MustCompile(`^Dialogue:.*?(\d+:\d+:\d+\.\d+),(\d+:\d+:\d+\.\d+),.*?,.*?,.*?,.*?,.*?,.*?,.*?,(.*)$`)

// assToVTT does a best-effort conversion from ASS/SSA to WebVTT.
// It strips override tags and extracts plain cue timings and text.
func assToVTT(ass []byte) string {
	var out strings.Builder
	out.WriteString("WEBVTT\n\n")
	scanner := bufio.NewScanner(bytes.NewReader(ass))
	for scanner.Scan() {
		line := scanner.Text()
		m := assTimingRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		start := strings.Replace(m[1], ".", ",", 1)
		end := strings.Replace(m[2], ".", ",", 1)
		// convert H:MM:SS.cc → HH:MM:SS.ccc
		start = assTimeToCue(start)
		end = assTimeToCue(end)
		text := stripAssTags(m[3])
		if text == "" {
			continue
		}
		fmt.Fprintf(&out, "%s --> %s\n%s\n\n", start, end, text)
	}
	return out.String()
}

func assTimeToCue(t string) string {
	// ASS uses H:MM:SS.cc (centiseconds); VTT uses HH:MM:SS.mmm
	parts := strings.SplitN(t, ":", 3)
	if len(parts) != 3 {
		return t
	}
	h := parts[0]
	if len(h) < 2 {
		h = "0" + h
	}
	secCs := strings.SplitN(parts[2], ".", 2)
	sec := secCs[0]
	cs := "00"
	if len(secCs) == 2 {
		cs = secCs[1]
	}
	ms := cs + "0"
	if len(ms) > 3 {
		ms = ms[:3]
	}
	return fmt.Sprintf("%s:%s:%s.%s", h, parts[1], sec, ms)
}

var assTagRe = regexp.MustCompile(`\{[^}]*\}`)

func stripAssTags(s string) string {
	s = assTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, `\N`, "\n")
	s = strings.ReplaceAll(s, `\n`, "\n")
	return strings.TrimSpace(s)
}

// langMap maps common subtitle filename tokens to ISO 639-1 codes.
var langMap = map[string]string{
	"en": "en", "eng": "en", "english": "en",
	"fr": "fr", "fra": "fr", "fre": "fr", "french": "fr", "francais": "fr", "français": "fr",
	"es": "es", "spa": "es", "spanish": "es", "espanol": "es", "español": "es",
	"de": "de", "deu": "de", "ger": "de", "german": "de", "deutsch": "de",
	"it": "it", "ita": "it", "italian": "it", "italiano": "it",
	"pt": "pt", "por": "pt", "portuguese": "pt", "portugues": "pt",
	"ru": "ru", "rus": "ru", "russian": "ru",
	"ar": "ar", "ara": "ar", "arabic": "ar",
	"zh": "zh", "zho": "zh", "chi": "zh", "chinese": "zh",
	"ja": "ja", "jpn": "ja", "japanese": "ja",
	"ko": "ko", "kor": "ko", "korean": "ko",
	"nl": "nl", "nld": "nl", "dut": "nl", "dutch": "nl",
	"pl": "pl", "pol": "pl", "polish": "pl",
	"tr": "tr", "tur": "tr", "turkish": "tr",
	"sv": "sv", "swe": "sv", "swedish": "sv",
	"no": "no", "nor": "no", "norwegian": "no",
	"da": "da", "dan": "da", "danish": "da",
	"fi": "fi", "fin": "fi", "finnish": "fi",
	"cs": "cs", "ces": "cs", "cze": "cs", "czech": "cs",
	"hu": "hu", "hun": "hu", "hungarian": "hu",
	"ro": "ro", "ron": "ro", "rum": "ro", "romanian": "ro",
	"he": "he", "heb": "he", "hebrew": "he",
	"hi": "hi", "hin": "hi", "hindi": "hi",
}

// detectLangFromPath tries to extract an ISO 639-1 language code from a
// subtitle file path. It checks each dot-separated and slash-separated
// segment of the name against langMap.
func detectLangFromPath(path string) string {
	base := strings.ToLower(filepath.Base(path))
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// Check parent directory name too (e.g. "Subs/English/movie.srt")
	dir := strings.ToLower(filepath.Base(filepath.Dir(path)))

	for _, token := range append(strings.FieldsFunc(base, func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == ' '
	}), dir) {
		if code, ok := langMap[token]; ok {
			return code
		}
	}
	return ""
}

var subtitleExts = map[string]bool{
	".srt": true, ".vtt": true, ".ass": true, ".ssa": true, ".sub": true,
}

// ExtractTorrentSubtitles scans torrent files for subtitle tracks, converts
// them to WebVTT, and saves them to the subtitle cache for the given TMDB ID.
// Already-cached files are not overwritten.
func ExtractTorrentSubtitles(t *torrent.Torrent, tmdbID int) {
	var subtitleFiles []*torrent.File
	for _, f := range t.Files() {
		f := f
		ext := strings.ToLower(filepath.Ext(f.DisplayPath()))
		if subtitleExts[ext] {
			subtitleFiles = append(subtitleFiles, f)
		}
	}
	if len(subtitleFiles) == 0 {
		return
	}

	// Count distinct detected languages to handle the single-file "und" case.
	for _, f := range subtitleFiles {
		lang := detectLangFromPath(f.DisplayPath())
		if lang == "" {
			if len(subtitleFiles) == 1 {
				lang = "und"
			} else {
				continue
			}
		}

		cachePath := subtitleCachePath(tmdbID, lang)
		if _, err := os.Stat(cachePath); err == nil {
			continue // already cached
		}

		r := f.NewReader()
		raw, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			log.Printf("subtitle: torrent extract read error %s: %v", f.DisplayPath(), err)
			continue
		}

		ext := strings.ToLower(filepath.Ext(f.DisplayPath()))
		var vtt string
		switch ext {
		case ".vtt":
			vtt = string(raw)
		case ".ass", ".ssa":
			vtt = assToVTT(raw)
		default: // .srt, .sub
			vtt = srtToVTT(raw)
		}

		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
			log.Printf("subtitle: torrent extract mkdir error: %v", err)
			continue
		}
		if err := os.WriteFile(cachePath, []byte(vtt), 0644); err != nil {
			log.Printf("subtitle: torrent extract write error: %v", err)
			continue
		}
		log.Printf("subtitle: extracted %s → %s (lang=%s)", f.DisplayPath(), cachePath, lang)
	}
}
