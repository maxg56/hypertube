package models

type CastMember struct {
	Name      string `json:"name"`
	Character string `json:"character"`
	Order     int    `json:"order"`
}

type Movie struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Year        string       `json:"year"`
	Overview    string       `json:"overview"`
	Runtime     int          `json:"runtime"`
	Rating      float64      `json:"rating"`
	PosterURL   string       `json:"poster_url"`
	BackdropURL string       `json:"backdrop_url"`
	Genres      []string     `json:"genres"`
	Cast        []CastMember `json:"cast"`
	Source      string       `json:"source"`
}

type SearchResult struct {
	Page       int     `json:"page"`
	TotalPages int     `json:"total_pages"`
	Results    []Movie `json:"results"`
}
