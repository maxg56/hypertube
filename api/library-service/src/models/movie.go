package models

import "time"

type CastMember struct {
	Name      string `json:"name"`
	Character string `json:"character"`
	Order     int    `json:"order"`
}

type Torrent struct {
	URL     string `json:"url"`
	Hash    string `json:"hash"`
	Quality string `json:"quality"`
	Type    string `json:"type"`
	Size    string `json:"size"`
	Seeds   int    `json:"seeds"`
	Peers   int    `json:"peers"`
	Magnet  string `json:"magnet,omitempty"`
}

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Movie struct {
	ID          int          `json:"id"`
	IMDbID      string       `json:"imdb_id,omitempty"`
	Title       string       `json:"title"`
	Year        string       `json:"year"`
	ReleaseDate string       `json:"release_date,omitempty"`
	Overview    string       `json:"overview"`
	Runtime     int          `json:"runtime"`
	Rating      float64      `json:"rating"`
	PosterURL   string       `json:"poster_url"`
	BackdropURL string       `json:"backdrop_url"`
	Genres      []string     `json:"genres"`
	Cast        []CastMember `json:"cast"`
	Torrents    []Torrent    `json:"torrents,omitempty"`
	Comments    []Comment    `json:"comments,omitempty"`
	Source      string       `json:"source"`
	Watched     bool         `json:"watched"`
}

type SearchResult struct {
	Page       int     `json:"page"`
	TotalPages int     `json:"total_pages"`
	TotalCount int     `json:"total_count"`
	Results    []Movie `json:"results"`
}

type CursorResult struct {
	Results    []Movie `json:"results"`
	NextCursor string  `json:"next_cursor,omitempty"`
	Total      int     `json:"total"`
}
