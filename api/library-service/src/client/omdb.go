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

const omdbBaseURL = "https://www.omdbapi.com"

type OMDbClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewOMDbClient() *OMDbClient {
	return &OMDbClient{
		apiKey:     os.Getenv("OMDB_API_KEY"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *OMDbClient) Available() bool {
	return c.apiKey != ""
}

type omdbDetailResponse struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Plot       string `json:"Plot"`
	Runtime    string `json:"Runtime"`
	ImdbRating string `json:"imdbRating"`
	Poster     string `json:"Poster"`
	Actors     string `json:"Actors"`
	Genre      string `json:"Genre"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

type omdbSearchResponse struct {
	Search       []omdbSearchItem `json:"Search"`
	TotalResults string           `json:"totalResults"`
	Response     string           `json:"Response"`
	Error        string           `json:"Error"`
}

type omdbSearchItem struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Poster string `json:"Poster"`
}

func (c *OMDbClient) Search(query string, page int) (*models.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("s", query)
	params.Set("type", "movie")
	params.Set("page", strconv.Itoa(page))

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/?%s", omdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw omdbSearchResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Response == "False" {
		return nil, fmt.Errorf("OMDb error: %s", raw.Error)
	}

	total, _ := strconv.Atoi(raw.TotalResults)
	totalPages := (total + 9) / 10

	result := &models.SearchResult{
		Page:       page,
		TotalPages: totalPages,
	}
	for _, r := range raw.Search {
		poster := r.Poster
		if poster == "N/A" {
			poster = ""
		}
		result.Results = append(result.Results, models.Movie{
			Title:     r.Title,
			Year:      r.Year,
			PosterURL: poster,
			Source:    "omdb",
		})
	}
	return result, nil
}

func (c *OMDbClient) GetMovieByTitle(title string) (*models.Movie, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("t", title)
	params.Set("type", "movie")
	params.Set("plot", "full")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/?%s", omdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw omdbDetailResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Response == "False" {
		return nil, fmt.Errorf("OMDb error: %s", raw.Error)
	}

	movie := &models.Movie{
		Title:    raw.Title,
		Year:     raw.Year,
		Overview: raw.Plot,
		Runtime:  parseRuntime(raw.Runtime),
		Rating:   parseRating(raw.ImdbRating),
		Source:   "omdb",
	}
	if raw.Poster != "N/A" {
		movie.PosterURL = raw.Poster
	}
	if raw.Genre != "N/A" && raw.Genre != "" {
		for _, g := range splitCSV(raw.Genre) {
			movie.Genres = append(movie.Genres, g)
		}
	}
	if raw.Actors != "N/A" && raw.Actors != "" {
		for _, a := range splitCSV(raw.Actors) {
			movie.Cast = append(movie.Cast, models.CastMember{Name: a})
		}
	}
	return movie, nil
}

func parseRuntime(s string) int {
	if len(s) < 4 {
		return 0
	}
	n, _ := strconv.Atoi(s[:len(s)-4])
	return n
}

func parseRating(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func splitCSV(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			part := trimSpace(s[start:i])
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	if part := trimSpace(s[start:]); part != "" {
		result = append(result, part)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
