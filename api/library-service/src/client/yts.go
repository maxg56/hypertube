package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"library-service/src/models"
)

const ytsBaseURL = "https://movies-api.accel.li/api/v2"

type YTSClient struct {
	httpClient *http.Client
}

func NewYTSClient() *YTSClient {
	return &YTSClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *YTSClient) Available() bool { return true }

// listResponse is used by list_movies.json (data.movies array)
type listResponse struct {
	Status string `json:"status"`
	Data   struct {
		MovieCount int        `json:"movie_count"`
		PageNumber int        `json:"page_number"`
		Limit      int        `json:"limit"`
		Movies     []ytsMovie `json:"movies"`
	} `json:"data"`
}

// detailResponse is used by movie_details.json (data.movie singular)
type detailResponse struct {
	Status string `json:"status"`
	Data   struct {
		Movie ytsMovieDetail `json:"movie"`
	} `json:"data"`
}

type ytsMovie struct {
	ID                  int          `json:"id"`
	Title               string       `json:"title"`
	Year                int          `json:"year"`
	Rating              float64      `json:"rating"`
	Runtime             int          `json:"runtime"`
	Genres              []string     `json:"genres"`
	Summary             string       `json:"summary"`
	IMDbCode            string       `json:"imdb_code"`
	MediumCoverImage    string       `json:"medium_cover_image"`
	BackgroundImageOrig string       `json:"background_image_original"`
	Torrents            []ytsTorrent `json:"torrents"`
}

type ytsMovieDetail struct {
	ID                  int          `json:"id"`
	Title               string       `json:"title"`
	Year                int          `json:"year"`
	Rating              float64      `json:"rating"`
	Runtime             int          `json:"runtime"`
	Genres              []string     `json:"genres"`
	DescriptionFull     string       `json:"description_full"`
	IMDbCode            string       `json:"imdb_code"`
	MediumCoverImage    string       `json:"medium_cover_image"`
	BackgroundImageOrig string       `json:"background_image_original"`
	Torrents            []ytsTorrent `json:"torrents"`
}

type ytsTorrent struct {
	URL     string `json:"url"`
	Hash    string `json:"hash"`
	Quality string `json:"quality"`
	Type    string `json:"type"`
	Size    string `json:"size"`
	Seeds   int    `json:"seeds"`
	Peers   int    `json:"peers"`
}

func (c *YTSClient) Search(query string, page int) (*models.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	params := url.Values{}
	params.Set("query_term", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("limit", "20")
	params.Set("sort_by", "seeds")

	body, err := c.get(fmt.Sprintf("%s/list_movies.json?%s", ytsBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}

	var raw listResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Status != "ok" {
		return nil, fmt.Errorf("YTS error: %s", raw.Status)
	}

	totalPages := 0
	if raw.Data.Limit > 0 {
		totalPages = (raw.Data.MovieCount + raw.Data.Limit - 1) / raw.Data.Limit
	}

	result := &models.SearchResult{
		Page:       page,
		TotalPages: totalPages,
	}
	for _, m := range raw.Data.Movies {
		result.Results = append(result.Results, listMovieToModel(m))
	}
	return result, nil
}

func (c *YTSClient) GetMovieByIMDbID(imdbID string) (*models.Movie, error) {
	params := url.Values{}
	params.Set("imdb_id", imdbID)
	params.Set("with_images", "true")

	body, err := c.get(fmt.Sprintf("%s/movie_details.json?%s", ytsBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}

	var raw detailResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Status != "ok" {
		return nil, fmt.Errorf("YTS error: %s", raw.Status)
	}
	if raw.Data.Movie.ID == 0 {
		return nil, nil
	}
	m := detailMovieToModel(raw.Data.Movie)
	return &m, nil
}

func (c *YTSClient) get(u string) ([]byte, error) {
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YTS returned status %d", resp.StatusCode)
	}
	return body, nil
}

func listMovieToModel(m ytsMovie) models.Movie {
	movie := models.Movie{
		ID:          m.ID,
		IMDbID:      m.IMDbCode,
		Title:       m.Title,
		Year:        strconv.Itoa(m.Year),
		Overview:    m.Summary,
		Runtime:     m.Runtime,
		Rating:      m.Rating,
		PosterURL:   m.MediumCoverImage,
		BackdropURL: m.BackgroundImageOrig,
		Genres:      m.Genres,
		Source:      "yts",
	}
	for _, t := range m.Torrents {
		movie.Torrents = append(movie.Torrents, toTorrent(t))
	}
	return movie
}

func detailMovieToModel(m ytsMovieDetail) models.Movie {
	movie := models.Movie{
		ID:          m.ID,
		IMDbID:      m.IMDbCode,
		Title:       m.Title,
		Year:        strconv.Itoa(m.Year),
		Overview:    m.DescriptionFull,
		Runtime:     m.Runtime,
		Rating:      m.Rating,
		PosterURL:   m.MediumCoverImage,
		BackdropURL: m.BackgroundImageOrig,
		Genres:      m.Genres,
		Source:      "yts",
	}
	for _, t := range m.Torrents {
		movie.Torrents = append(movie.Torrents, toTorrent(t))
	}
	return movie
}

func toTorrent(t ytsTorrent) models.Torrent {
	return models.Torrent{
		URL:     t.URL,
		Hash:    t.Hash,
		Quality: t.Quality,
		Type:    t.Type,
		Size:    t.Size,
		Seeds:   t.Seeds,
		Peers:   t.Peers,
	}
}
