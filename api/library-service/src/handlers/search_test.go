package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"library-service/src/models"
	"library-service/src/utils"
)

func TestSearch_MissingQuery(t *testing.T) {
	h := &MovieHandler{tmdb: &mockTMDb{available: true}}
	c, w := newCtx("GET", "/search")
	h.Search(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearch_TMDbReturnsResults(t *testing.T) {
	want := sampleResult("Inception", "Batman")
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, searchResult: want},
		omdb: &mockSearch{available: false},
	}
	c, w := newCtx("GET", "/search?q=inception")
	h.Search(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp utils.StandardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success=true")
	}
}

func TestSearch_FallsBackToOMDb(t *testing.T) {
	want := sampleResult("The Dark Knight")
	h := &MovieHandler{
		tmdb: &mockTMDb{available: false},
		omdb: &mockSearch{available: true, result: want},
	}
	c, w := newCtx("GET", "/search?q=batman")
	h.Search(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSearch_TMDbErrorFallsBackToOMDb(t *testing.T) {
	want := sampleResult("Interstellar")
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, searchErr: errors.New("timeout")},
		omdb: &mockSearch{available: true, result: want},
	}
	c, w := newCtx("GET", "/search?q=nolan")
	h.Search(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSearch_BothProvidersUnavailable(t *testing.T) {
	h := &MovieHandler{
		tmdb: &mockTMDb{available: false},
		omdb: &mockSearch{available: false},
	}
	c, w := newCtx("GET", "/search?q=batman")
	h.Search(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestSearch_BothProvidersError(t *testing.T) {
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, searchErr: errors.New("fail")},
		omdb: &mockSearch{available: true, err: errors.New("fail")},
	}
	c, w := newCtx("GET", "/search?q=batman")
	h.Search(c)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestSearchYTS_MissingQuery(t *testing.T) {
	h := &MovieHandler{yts: &mockYTS{}}
	c, w := newCtx("GET", "/yts")
	h.SearchYTS(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearchYTS_ReturnsResults(t *testing.T) {
	want := sampleResult("Avengers")
	h := &MovieHandler{yts: &mockYTS{searchResult: want}}
	c, w := newCtx("GET", "/yts?q=avengers")
	h.SearchYTS(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success=true")
	}
}

func TestSearchYTS_Error(t *testing.T) {
	h := &MovieHandler{yts: &mockYTS{searchErr: errors.New("YTS down")}}
	c, w := newCtx("GET", "/yts?q=test")
	h.SearchYTS(c)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestSearch_ResultShape(t *testing.T) {
	want := &models.SearchResult{
		Page:       2,
		TotalPages: 5,
		Results:    sampleMovies("Movie A", "Movie B"),
	}
	h := &MovieHandler{
		tmdb: &mockTMDb{available: true, searchResult: want},
		omdb: &mockSearch{},
	}
	c, w := newCtx("GET", "/search?q=test&page=2")
	h.Search(c)

	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, _ := json.Marshal(resp.Data)
	var result models.SearchResult
	json.Unmarshal(data, &result)

	if result.Page != 2 {
		t.Errorf("page=%d, want 2", result.Page)
	}
	if result.TotalPages != 5 {
		t.Errorf("total_pages=%d, want 5", result.TotalPages)
	}
	if len(result.Results) != 2 {
		t.Errorf("results count=%d, want 2", len(result.Results))
	}
}
