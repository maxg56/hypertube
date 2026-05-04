package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"library-service/src/models"
	"library-service/src/utils"
)

func TestGetMovie_InvalidID(t *testing.T) {
	h := &MovieHandler{tmdb: &mockTMDb{available: true}}
	c, w := newCtx("GET", "/movies/abc")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	h.GetMovie(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetMovie_TMDbUnavailable(t *testing.T) {
	h := &MovieHandler{tmdb: &mockTMDb{available: false}}
	c, w := newCtx("GET", "/movies/1")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetMovie(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestGetMovie_NotFound(t *testing.T) {
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, movieResult: nil},
		yts:  &mockYTS{},
	}
	c, w := newCtx("GET", "/movies/999")
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	h.GetMovie(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetMovie_TMDbError(t *testing.T) {
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, movieErr: errors.New("timeout")},
		yts:  &mockYTS{},
	}
	c, w := newCtx("GET", "/movies/1")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetMovie(c)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestGetMovie_FoundWithoutTorrents(t *testing.T) {
	movie := &models.Movie{ID: 42, Title: "Inception", IMDbID: ""}
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, movieResult: movie},
		yts:  &mockYTS{},
	}
	c, w := newCtx("GET", "/movies/42")
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	h.GetMovie(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success=true")
	}
}

func TestGetMovie_EnrichedWithYTSTorrents(t *testing.T) {
	movie := &models.Movie{ID: 1, Title: "The Matrix", IMDbID: "tt0133093"}
	torrents := []models.Torrent{{Hash: "abc123", Quality: "1080p", Seeds: 100}}
	ytsMovie := &models.Movie{Torrents: torrents}

	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, movieResult: movie},
		yts:  &mockYTS{imdbResult: ytsMovie},
	}
	c, w := newCtx("GET", "/movies/1")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetMovie(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := json.Marshal(resp.Data)
	var result models.Movie
	json.Unmarshal(data, &result)

	if len(result.Torrents) != 1 {
		t.Errorf("expected 1 torrent, got %d", len(result.Torrents))
	}
	if result.Torrents[0].Quality != "1080p" {
		t.Errorf("expected quality=1080p, got %q", result.Torrents[0].Quality)
	}
}

func TestGetMovie_YTSErrorDoesNotFail(t *testing.T) {
	movie := &models.Movie{ID: 1, Title: "Dune", IMDbID: "tt1160419"}
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, movieResult: movie},
		yts:  &mockYTS{imdbErr: errors.New("YTS down")},
	}
	c, w := newCtx("GET", "/movies/1")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.GetMovie(c)

	if w.Code != http.StatusOK {
		t.Errorf("YTS error should not block movie response, got %d", w.Code)
	}
}
