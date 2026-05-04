package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"library-service/src/models"
)

const (
	tmdbBaseURL   = "https://api.themoviedb.org/3"
	tmdbImageBase = "https://image.tmdb.org/t/p/w500"
)

type TMDbClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewTMDbClient() *TMDbClient {
	return &TMDbClient{
		apiKey:     os.Getenv("TMDB_API_KEY"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TMDbClient) Available() bool {
	return c.apiKey != ""
}

type tmdbSearchResponse struct {
	Page         int               `json:"page"`
	TotalPages   int               `json:"total_pages"`
	Results      []tmdbMovieResult `json:"results"`
}

type tmdbMovieResult struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	ReleaseDate      string  `json:"release_date"`
	Overview         string  `json:"overview"`
	VoteAverage      float64 `json:"vote_average"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
}

type tmdbDetailResponse struct {
	ID           int     `json:"id"`
	IMDbID       string  `json:"imdb_id"`
	Title        string  `json:"title"`
	ReleaseDate  string  `json:"release_date"`
	Overview     string  `json:"overview"`
	Runtime      int     `json:"runtime"`
	VoteAverage  float64 `json:"vote_average"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	Genres       []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Credits struct {
		Cast []struct {
			Name      string `json:"name"`
			Character string `json:"character"`
			Order     int    `json:"order"`
		} `json:"cast"`
	} `json:"credits"`
}

func (c *TMDbClient) Search(query string, page int) (*models.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search/movie?%s", tmdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb search returned status %d", resp.StatusCode)
	}

	var raw tmdbSearchResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	result := &models.SearchResult{
		Page:       raw.Page,
		TotalPages: raw.TotalPages,
	}
	for _, r := range raw.Results {
		result.Results = append(result.Results, models.Movie{
			ID:          r.ID,
			Title:       r.Title,
			Year:        releaseYear(r.ReleaseDate),
			Overview:    r.Overview,
			Rating:      r.VoteAverage,
			PosterURL:   imageURL(r.PosterPath),
			BackdropURL: imageURL(r.BackdropPath),
			Source:      "tmdb",
		})
	}
	return result, nil
}

func (c *TMDbClient) GetMovie(id int) (*models.Movie, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("append_to_response", "credits")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/movie/%d?%s", tmdbBaseURL, id, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb detail returned status %d", resp.StatusCode)
	}

	var raw tmdbDetailResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	movie := &models.Movie{
		ID:          raw.ID,
		IMDbID:      raw.IMDbID,
		Title:       raw.Title,
		Year:        releaseYear(raw.ReleaseDate),
		ReleaseDate: raw.ReleaseDate,
		Overview:    raw.Overview,
		Runtime:     raw.Runtime,
		Rating:      raw.VoteAverage,
		PosterURL:   imageURL(raw.PosterPath),
		BackdropURL: imageURL(raw.BackdropPath),
		Source:      "tmdb",
	}
	for _, g := range raw.Genres {
		movie.Genres = append(movie.Genres, g.Name)
	}
	for i, m := range raw.Credits.Cast {
		if i >= 10 {
			break
		}
		movie.Cast = append(movie.Cast, models.CastMember{
			Name:      m.Name,
			Character: m.Character,
			Order:     m.Order,
		})
	}
	return movie, nil
}

func imageURL(path string) string {
	if path == "" {
		return ""
	}
	return tmdbImageBase + path
}

func releaseYear(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return date
}
