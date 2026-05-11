package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"torrent-service/src/conf"
	"torrent-service/src/handlers"
	"torrent-service/src/models"
)

const (
	validMagnet   = "magnet:?xt=urn:btih:da39a3ee5e6b4b0d3255bfef95601890afd80709&dn=test"
	validInfoHash = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE torrents (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id   INTEGER,
		magnet_uri TEXT    NOT NULL,
		info_hash  TEXT    UNIQUE NOT NULL,
		status     TEXT    DEFAULT 'pending',
		file_path  TEXT,
		file_size  INTEGER DEFAULT 0,
		downloaded INTEGER DEFAULT 0,
		progress   REAL    DEFAULT 0,
		quality    TEXT,
			source     TEXT,
		error_msg  TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error)
	conf.DB = db
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "torrent-service"})
	})
	api := r.Group("/api/v1")
	api.POST("/torrent/download", handlers.DownloadHandler)
	api.GET("/torrent/status/:id", handlers.StatusHandler)
	api.GET("/stream/:id", handlers.StreamHandler)
	return r
}

func post(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	return m
}

// ---- /health ----

func TestHealthCheck(t *testing.T) {
	r := newRouter()
	w := get(r, "/health")
	assert.Equal(t, http.StatusOK, w.Code)
	body := decode(t, w)
	assert.Equal(t, "torrent-service", body["service"])
	assert.Equal(t, "ok", body["status"])
}

// ---- POST /api/v1/torrent/download ----

func TestDownloadHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErrMsg string
	}{
		{
			name:       "missing body",
			body:       nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing movie_id",
			body:       map[string]any{"magnet_uri": validMagnet},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing magnet_uri",
			body:       map[string]any{"movie_id": 1},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid magnet URI",
			body:       map[string]any{"magnet_uri": "not-a-magnet", "movie_id": 1},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "invalid magnet URI",
		},
		{
			// torrent client is nil in tests → StartDownload returns an error,
			// but the handler converts it to 400 for all service-level errors.
			name:       "valid magnet but client not initialized",
			body:       map[string]any{"magnet_uri": validMagnet, "movie_id": 1},
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "torrent client not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDB(t)
			r := newRouter()
			w := post(r, "/api/v1/torrent/download", tt.body)
			assert.Equal(t, tt.wantStatus, w.Code)
			body := decode(t, w)
			assert.Equal(t, false, body["success"])
			if tt.wantErrMsg != "" {
				assert.Contains(t, body["error"], tt.wantErrMsg)
			}
		})
	}
}

// ---- GET /api/v1/torrent/status/:id ----

func TestStatusHandler_NotFound(t *testing.T) {
	setupTestDB(t)
	r := newRouter()
	w := get(r, "/api/v1/torrent/status/unknownhash")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, false, decode(t, w)["success"])
}

func TestStatusHandler_Statuses(t *testing.T) {
	statuses := []models.TorrentStatus{
		models.StatusPending,
		models.StatusDownloading,
		models.StatusReady,
		models.StatusError,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			setupTestDB(t)
			conf.DB.Create(&models.TorrentRecord{
				InfoHash:  validInfoHash,
				MagnetURI: validMagnet,
				MovieID:   1,
				Status:    status,
				Progress:  42.5,
			})

			r := newRouter()
			w := get(r, "/api/v1/torrent/status/"+validInfoHash)
			assert.Equal(t, http.StatusOK, w.Code)

			body := decode(t, w)
			assert.Equal(t, true, body["success"])
			data := body["data"].(map[string]any)
			assert.Equal(t, string(status), data["status"])
			assert.Equal(t, validInfoHash, data["info_hash"])
			assert.InDelta(t, 42.5, data["progress"].(float64), 0.01)
		})
	}
}

// ---- GET /api/v1/stream/:id ----

func TestStreamHandler_NotFound(t *testing.T) {
	setupTestDB(t)
	r := newRouter()
	w := get(r, "/api/v1/stream/unknownhash")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStreamHandler_Pending(t *testing.T) {
	setupTestDB(t)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusPending,
	})
	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "5", w.Header().Get("Retry-After"))
}

func TestStreamHandler_Error(t *testing.T) {
	setupTestDB(t)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusError,
		ErrorMsg:  "no seeders",
	})
	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	body := decode(t, w)
	assert.Contains(t, body["error"], "no seeders")
}

func TestStreamHandler_Downloading_ClientNil(t *testing.T) {
	// Torrent is downloading but not in activeTorrents (no client) → 503.
	setupTestDB(t)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusDownloading,
	})
	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestStreamHandler_Ready_FileMissing(t *testing.T) {
	// Torrent is marked ready but the file no longer exists on disk → 503.
	setupTestDB(t)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  "/tmp/nonexistent_test_file_xyz.mkv",
		FileSize:  1024,
	})
	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// makeTempVideoFile creates a temp file with the given extension and content,
// registers a StatusReady DB record for it, and returns the file path.
func makeTempVideoFile(t *testing.T, ext string, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "testvideo*"+ext)
	require.NoError(t, err)
	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestStreamHandler_Ready_Success(t *testing.T) {
	setupTestDB(t)
	content := bytes.Repeat([]byte("A"), 100)
	path := makeTempVideoFile(t, ".mp4", content)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  path,
		FileSize:  int64(len(content)),
	})
	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
	assert.Equal(t, len(content), w.Body.Len())
}

func TestStreamHandler_Range_PartialContent(t *testing.T) {
	setupTestDB(t)
	content := make([]byte, 100)
	for i := range content {
		content[i] = byte(i)
	}
	path := makeTempVideoFile(t, ".mp4", content)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  path,
		FileSize:  int64(len(content)),
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stream/"+validInfoHash, nil)
	req.Header.Set("Range", "bytes=0-9")
	w := httptest.NewRecorder()
	newRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "bytes 0-9/100", w.Header().Get("Content-Range"))
	assert.Equal(t, 10, w.Body.Len())
	assert.Equal(t, content[:10], w.Body.Bytes())
}

func TestStreamHandler_MimeType(t *testing.T) {
	// Only native formats go through http.ServeContent which infers MIME from extension.
	// Non-native formats (mkv, avi, …) are transcoded and always return video/mp4.
	tests := []struct {
		ext      string
		wantMime string
	}{
		{".mp4", "video/mp4"},
		{".webm", "video/webm"},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			setupTestDB(t)
			path := makeTempVideoFile(t, tt.ext, bytes.Repeat([]byte("B"), 20))
			conf.DB.Create(&models.TorrentRecord{
				InfoHash:  validInfoHash,
				MagnetURI: validMagnet,
				MovieID:   1,
				Status:    models.StatusReady,
				FilePath:  path,
				FileSize:  20,
			})
			r := newRouter()
			w := get(r, "/api/v1/stream/"+validInfoHash)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), tt.wantMime)
		})
	}
}

// makeTestMKVFile creates a real 1-second H.264/AAC MKV via ffmpeg and registers
// it in the DB as StatusReady. Skips the test if ffmpeg is not in PATH.
func makeTestMKVFile(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	out := filepath.Join(t.TempDir(), "test.mkv")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "testsrc=duration=1:size=64x64:rate=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-c:v", "libx264", "-c:a", "aac",
		"-t", "1", "-y", out,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "ffmpeg test-video generation failed: %s", output)
	return out
}

func TestStreamHandler_Transcode_ContentType(t *testing.T) {
	setupTestDB(t)
	path := makeTestMKVFile(t)

	info, _ := os.Stat(path)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  path,
		FileSize:  info.Size(),
	})

	r := newRouter()
	w := get(r, "/api/v1/stream/"+validInfoHash)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "video/mp4")
	assert.Greater(t, w.Body.Len(), 64, "transcoded output should contain MP4 data")
}

func TestStreamHandler_Transcode_NoByteRange(t *testing.T) {
	// Transcoded streams do not support byte-range (ffmpeg pipe is not seekable).
	setupTestDB(t)
	path := makeTestMKVFile(t)

	info, _ := os.Stat(path)
	conf.DB.Create(&models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  path,
		FileSize:  info.Size(),
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stream/"+validInfoHash, nil)
	req.Header.Set("Range", "bytes=0-99")
	w := httptest.NewRecorder()
	newRouter().ServeHTTP(w, req)

	// Transcoded response is always 200 (full stream), never 206 Partial Content.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Accept-Ranges"))
}
