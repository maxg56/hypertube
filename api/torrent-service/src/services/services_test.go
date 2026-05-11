package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"torrent-service/src/conf"
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
	require.NoError(t, db.Exec(`CREATE TABLE movies (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		tmdb_id     INTEGER UNIQUE NOT NULL,
		title       TEXT    NOT NULL,
		poster_path TEXT
	)`).Error)
	conf.DB = db
}

// ---- extractInfoHash ----

func TestExtractInfoHash(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantHash string
		wantErr  bool
	}{
		{
			name:     "valid magnet",
			uri:      validMagnet,
			wantHash: validInfoHash,
		},
		{
			name:    "http url not supported",
			uri:     "https://example.com/file.torrent",
			wantErr: true,
		},
		{
			name:    "empty string",
			uri:     "",
			wantErr: true,
		},
		{
			// anacrolix parses this without error and returns a zero hash —
			// the library accepts magnets without btih as "no known hash".
			name:     "magnet without btih returns zero hash",
			uri:      "magnet:?dn=noHash",
			wantHash: "0000000000000000000000000000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractInfoHash(tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantHash, got)
			}
		})
	}
}

func TestExtractInfoHashIsLowercase(t *testing.T) {
	upper := "magnet:?xt=urn:btih:DA39A3EE5E6B4B0D3255BFEF95601890AFD80709&dn=test"
	hash, err := extractInfoHash(upper)
	require.NoError(t, err)
	assert.Equal(t, validInfoHash, hash)
}

// ---- findOrCreateRecord ----

func TestFindOrCreateRecord_NewRecord(t *testing.T) {
	setupTestDB(t)

	rec, err := findOrCreateRecord(validMagnet, validInfoHash, 1, "")
	require.NoError(t, err)
	assert.Equal(t, validInfoHash, rec.InfoHash)
	assert.Equal(t, validMagnet, rec.MagnetURI)
	assert.Equal(t, 1, rec.MovieID)
	assert.Equal(t, models.StatusPending, rec.Status)
}

func TestFindOrCreateRecord_ExistingReady(t *testing.T) {
	setupTestDB(t)

	seed := models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusReady,
		FilePath:  "/data/torrents/test.mkv",
	}
	require.NoError(t, conf.DB.Create(&seed).Error)

	// Simulate file existing on disk by using a path we know won't exist —
	// findOrCreateRecord only checks os.Stat when status==ready, and falls
	// through to re-download if the file is missing.
	rec, err := findOrCreateRecord(validMagnet, validInfoHash, 1, "")
	require.NoError(t, err)
	// File doesn't exist on disk → record is reset to pending for re-download.
	assert.Equal(t, models.StatusPending, rec.Status)
}

func TestFindOrCreateRecord_ExistingDownloading(t *testing.T) {
	setupTestDB(t)

	seed := models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusDownloading,
	}
	require.NoError(t, conf.DB.Create(&seed).Error)

	rec, err := findOrCreateRecord(validMagnet, validInfoHash, 1, "")
	require.NoError(t, err)
	assert.Equal(t, models.StatusDownloading, rec.Status)
}

func TestFindOrCreateRecord_ExistingPending(t *testing.T) {
	setupTestDB(t)

	seed := models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusPending,
	}
	require.NoError(t, conf.DB.Create(&seed).Error)

	rec, err := findOrCreateRecord(validMagnet, validInfoHash, 1, "")
	require.NoError(t, err)
	assert.Equal(t, models.StatusPending, rec.Status)
}

func TestFindOrCreateRecord_ErrorStatusReset(t *testing.T) {
	setupTestDB(t)

	seed := models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusError,
		ErrorMsg:  "previous failure",
	}
	require.NoError(t, conf.DB.Create(&seed).Error)

	rec, err := findOrCreateRecord(validMagnet, validInfoHash, 1, "")
	require.NoError(t, err)
	assert.Equal(t, models.StatusPending, rec.Status)
	assert.Empty(t, rec.ErrorMsg)
}

// ---- GetRecord ----

func TestGetRecord_NotFound(t *testing.T) {
	setupTestDB(t)

	_, err := GetRecord("nonexistenthash")
	assert.Error(t, err)
}

func TestGetRecord_Found(t *testing.T) {
	setupTestDB(t)

	seed := models.TorrentRecord{
		InfoHash:  validInfoHash,
		MagnetURI: validMagnet,
		MovieID:   1,
		Status:    models.StatusDownloading,
	}
	require.NoError(t, conf.DB.Create(&seed).Error)

	rec, err := GetRecord(validInfoHash)
	require.NoError(t, err)
	assert.Equal(t, validInfoHash, rec.InfoHash)
	assert.Equal(t, models.StatusDownloading, rec.Status)
}

// ---- StartDownload ----

func TestStartDownload_ClientNotInitialized(t *testing.T) {
	setupTestDB(t)
	client = nil // ensure no client

	_, err := StartDownload(validMagnet, 1, "")
	assert.EqualError(t, err, "torrent client not initialized")
}

func TestStartDownload_InvalidMagnet(t *testing.T) {
	setupTestDB(t)
	client = nil

	_, err := StartDownload("not-a-magnet", 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid magnet URI")
}
