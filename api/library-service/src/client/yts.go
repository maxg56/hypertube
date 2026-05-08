package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
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

// ListParams holds optional filters and sort options for List.
type ListParams struct {
	Query     string
	Genre     string
	MinRating float64
	Year      int    // filtered client-side (YTS has no year param)
	SortBy    string // "seeds" | "title" | "year" | "rating"
	Page      int
}

func (c *YTSClient) Search(query string, page int) (*models.SearchResult, error) {
	return c.List(ListParams{Query: query, Page: page, SortBy: "seeds"})
}

// List fetches movies from YTS with optional filters and sorting.
// When a year filter is active, it scans up to maxYearScanPages YTS pages
// starting from p.Page to collect matching results, because YTS has no
// native year parameter.
func (c *YTSClient) List(p ListParams) (*models.SearchResult, error) {
	if p.Page < 1 {
		p.Page = 1
	}

	validSort := map[string]bool{"seeds": true, "title": true, "year": true, "rating": true, "download_count": true}
	sortBy := "seeds"
	if validSort[p.SortBy] {
		sortBy = p.SortBy
	}

	baseParams := url.Values{}
	if p.Query != "" {
		baseParams.Set("query_term", p.Query)
	}
	if p.Genre != "" {
		baseParams.Set("genre", p.Genre)
	}
	if p.MinRating > 0 {
		baseParams.Set("minimum_rating", strconv.FormatFloat(p.MinRating, 'f', 0, 64))
	}
	baseParams.Set("sort_by", sortBy)
	baseParams.Set("limit", "20")

	// Fast path: no year filter — single YTS request.
	if p.Year == 0 {
		baseParams.Set("page", strconv.Itoa(p.Page))
		body, err := c.get(fmt.Sprintf("%s/list_movies.json?%s", ytsBaseURL, baseParams.Encode()))
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
			Page:       p.Page,
			TotalPages: totalPages,
			TotalCount: raw.Data.MovieCount,
		}
		for _, m := range raw.Data.Movies {
			result.Results = append(result.Results, listMovieToModel(m))
		}
		return result, nil
	}

	// Year filter: fire maxYearScanPages YTS requests in parallel, filter by
	// year, and return all matches. The cursor advances by maxYearScanPages so
	// successive "load more" calls don't re-fetch the same pages.
	const maxYearScanPages = 10

	type pageResult struct {
		movies     []ytsMovie
		totalPages int
		empty      bool
	}

	results := make([]pageResult, maxYearScanPages)
	var wg sync.WaitGroup
	for i := 0; i < maxYearScanPages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			params := url.Values{}
			for k, v := range baseParams {
				params[k] = v
			}
			params.Set("page", strconv.Itoa(p.Page+idx))

			body, err := c.get(fmt.Sprintf("%s/list_movies.json?%s", ytsBaseURL, params.Encode()))
			if err != nil {
				results[idx].empty = true
				return
			}
			var raw listResponse
			if err := json.Unmarshal(body, &raw); err != nil || raw.Status != "ok" || len(raw.Data.Movies) == 0 {
				results[idx].empty = true
				return
			}
			tp := 0
			if raw.Data.Limit > 0 {
				tp = (raw.Data.MovieCount + raw.Data.Limit - 1) / raw.Data.Limit
			}
			results[idx] = pageResult{movies: raw.Data.Movies, totalPages: tp}
		}(i)
	}
	wg.Wait()

	totalYTSPages := 0
	collected := make([]models.Movie, 0, 40)
	for _, r := range results {
		if r.empty {
			continue
		}
		if r.totalPages > totalYTSPages {
			totalYTSPages = r.totalPages
		}
		for _, m := range r.movies {
			if m.Year == p.Year {
				collected = append(collected, listMovieToModel(m))
			}
		}
	}

	nextYTSPage := p.Page + maxYearScanPages
	hasMore := totalYTSPages == 0 || nextYTSPage <= totalYTSPages
	result := &models.SearchResult{
		Page:       p.Page,
		TotalPages: p.Page + 1,
		TotalCount: len(collected),
	}
	if !hasMore {
		result.TotalPages = p.Page
	}
	result.Results = collected
	return result, nil
}

func (c *YTSClient) GetMovieByID(id int) (*models.Movie, error) {
	params := url.Values{}
	params.Set("movie_id", strconv.Itoa(id))
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
