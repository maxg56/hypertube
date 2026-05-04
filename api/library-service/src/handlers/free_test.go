package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"library-service/src/models"
	"library-service/src/utils"
)

func TestDeduplicateMovies_NoDuplicates(t *testing.T) {
	a := freeResult{result: sampleResult("Inception", "Batman")}
	b := freeResult{result: sampleResult("Dune", "Arrival")}
	got := deduplicateMovies(a, b)
	if len(got) != 4 {
		t.Errorf("expected 4 movies, got %d", len(got))
	}
}

func TestDeduplicateMovies_RemovesDuplicates(t *testing.T) {
	a := freeResult{result: sampleResult("Batman", "Inception")}
	b := freeResult{result: sampleResult("Batman", "Dune")} // "Batman" duplicated
	got := deduplicateMovies(a, b)
	if len(got) != 3 {
		t.Errorf("expected 3 movies, got %d: %v", len(got), titlesOf(got))
	}
}

func TestDeduplicateMovies_CaseInsensitive(t *testing.T) {
	a := freeResult{result: sampleResult("star wars")}
	b := freeResult{result: sampleResult("Star Wars")}
	got := deduplicateMovies(a, b)
	if len(got) != 1 {
		t.Errorf("expected 1 movie after case-insensitive dedup, got %d", len(got))
	}
}

func TestDeduplicateMovies_PunctuationIgnored(t *testing.T) {
	a := freeResult{result: sampleResult("Star Wars: Episode IV")}
	b := freeResult{result: sampleResult("Star Wars Episode IV")}
	got := deduplicateMovies(a, b)
	if len(got) != 1 {
		t.Errorf("expected 1 movie after punctuation-insensitive dedup, got %d", len(got))
	}
}

func TestDeduplicateMovies_SkipsErrorSources(t *testing.T) {
	a := freeResult{err: errors.New("down"), result: nil}
	b := freeResult{result: sampleResult("Dune")}
	got := deduplicateMovies(a, b)
	if len(got) != 1 {
		t.Errorf("expected 1 movie from healthy source, got %d", len(got))
	}
}

func TestDeduplicateMovies_SkipsEmptyTitle(t *testing.T) {
	a := freeResult{result: &models.SearchResult{
		Results: []models.Movie{{Title: ""}, {Title: "Dune"}},
	}}
	got := deduplicateMovies(a)
	if len(got) != 1 {
		t.Errorf("expected 1 movie (empty title skipped), got %d", len(got))
	}
}

func TestSearchFree_MissingQuery(t *testing.T) {
	h := &MovieHandler{legit: &mockFree{}, archive: &mockFree{}}
	c, w := newCtx("GET", "/free")
	h.SearchFree(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearchFree_BothSourcesError(t *testing.T) {
	h := &MovieHandler{
		legit:   &mockFree{err: errors.New("down")},
		archive: &mockFree{err: errors.New("down")},
	}
	c, w := newCtx("GET", "/free?q=batman")
	h.SearchFree(c)
	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestSearchFree_OneSourceError_StillResponds(t *testing.T) {
	h := &MovieHandler{
		legit:   &mockFree{err: errors.New("legit down")},
		archive: &mockFree{result: sampleResult("Dune")},
	}
	c, w := newCtx("GET", "/free?q=dune")
	h.SearchFree(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success=true")
	}
}

func TestSearchFree_MergesAndDeduplicates(t *testing.T) {
	h := &MovieHandler{
		legit:   &mockFree{result: sampleResult("Batman", "Inception")},
		archive: &mockFree{result: sampleResult("Batman", "Dune")},
	}
	c, w := newCtx("GET", "/free?q=batman")
	h.SearchFree(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp utils.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := json.Marshal(resp.Data)
	var result models.SearchResult
	json.Unmarshal(data, &result)

	if len(result.Results) != 3 {
		t.Errorf("expected 3 movies after dedup, got %d: %v", len(result.Results), titlesOf(result.Results))
	}
}

func titlesOf(movies []models.Movie) []string {
	titles := make([]string, len(movies))
	for i, m := range movies {
		titles[i] = m.Title
	}
	return titles
}
