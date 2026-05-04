package handlers

import (
	"net/http"
	"testing"
)

func TestParseQueryPage_MissingQuery(t *testing.T) {
	c, w := newCtx("GET", "/search")
	_, _, ok := parseQueryPage(c)
	if ok {
		t.Fatal("expected ok=false when q is missing")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestParseQueryPage_Valid(t *testing.T) {
	tests := []struct {
		url      string
		wantQ    string
		wantPage int
	}{
		{"/search?q=batman", "batman", 1},
		{"/search?q=inception&page=3", "inception", 3},
		{"/search?q=test&page=0", "test", 1},  // page<1 → 1
		{"/search?q=test&page=-5", "test", 1}, // page<1 → 1
	}
	for _, tt := range tests {
		c, _ := newCtx("GET", tt.url)
		q, page, ok := parseQueryPage(c)
		if !ok {
			t.Errorf("url=%s: expected ok=true", tt.url)
			continue
		}
		if q != tt.wantQ {
			t.Errorf("url=%s: query=%q, want %q", tt.url, q, tt.wantQ)
		}
		if page != tt.wantPage {
			t.Errorf("url=%s: page=%d, want %d", tt.url, page, tt.wantPage)
		}
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Batman Begins", "batmanbegins"},
		{"L'Épée du Roi", "lépéeduroi"},
		{"Star Wars: Episode IV", "starwarsepisodeiv"},
		{"2001: A Space Odyssey", "2001aspaceodyssey"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		got := normalizeTitle(tt.input)
		if got != tt.want {
			t.Errorf("normalizeTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
