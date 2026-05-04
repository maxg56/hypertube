package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"library-service/src/client"
	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

const (
	ytsCacheTTL  = 1 * time.Hour
	freeCacheTTL = 2 * time.Hour
)

type MovieHandler struct {
	tmdb   *client.TMDbClient
	omdb   *client.OMDbClient
	yts    *client.YTSClient
	legit  *client.LegitTorrentsClient
	archive *client.ArchiveClient
}

func NewMovieHandler() *MovieHandler {
	return &MovieHandler{
		tmdb:    client.NewTMDbClient(),
		omdb:    client.NewOMDbClient(),
		yts:     client.NewYTSClient(),
		legit:   client.NewLegitTorrentsClient(),
		archive: client.NewArchiveClient(),
	}
}

func (h *MovieHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.RespondError(c, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	cacheKey := fmt.Sprintf("search:%s:page:%d", query, page)
	if cached, err := conf.GetCache(cacheKey); err == nil {
		var result models.SearchResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			utils.RespondSuccess(c, http.StatusOK, result)
			return
		}
	}

	var result *models.SearchResult
	var err error

	if h.tmdb.Available() {
		result, err = h.tmdb.Search(query, page)
	}

	if (result == nil || err != nil) && h.omdb.Available() {
		result, err = h.omdb.Search(query, page)
	}

	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch search results")
		return
	}
	if result == nil {
		utils.RespondError(c, http.StatusServiceUnavailable, "no metadata provider available")
		return
	}

	if data, err := json.Marshal(result); err == nil {
		_ = conf.SetCache(cacheKey, string(data), conf.MovieCacheTTL)
	}

	utils.RespondSuccess(c, http.StatusOK, result)
}

func (h *MovieHandler) GetMovie(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	cacheKey := fmt.Sprintf("movie:%d", id)
	if cached, err := conf.GetCache(cacheKey); err == nil && err != redis.Nil {
		var movie models.Movie
		if json.Unmarshal([]byte(cached), &movie) == nil {
			utils.RespondSuccess(c, http.StatusOK, movie)
			return
		}
	}

	var movie *models.Movie

	if h.tmdb.Available() {
		movie, err = h.tmdb.GetMovie(id)
	}

	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch movie details")
		return
	}
	if movie == nil {
		utils.RespondError(c, http.StatusNotFound, "movie not found")
		return
	}

	if movie.IMDbID != "" {
		if ytsMovie, ytsErr := h.yts.GetMovieByIMDbID(movie.IMDbID); ytsErr == nil && ytsMovie != nil {
			movie.Torrents = ytsMovie.Torrents
		}
	}

	if data, err := json.Marshal(movie); err == nil {
		_ = conf.SetCache(cacheKey, string(data), conf.MovieCacheTTL)
	}

	utils.RespondSuccess(c, http.StatusOK, movie)
}

func (h *MovieHandler) SearchYTS(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.RespondError(c, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	cacheKey := fmt.Sprintf("yts:search:%s:page:%d", query, page)
	if cached, err := conf.GetCache(cacheKey); err == nil {
		var result models.SearchResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			utils.RespondSuccess(c, http.StatusOK, result)
			return
		}
	}

	result, err := h.yts.Search(query, page)
	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch YTS results")
		return
	}

	if data, err := json.Marshal(result); err == nil {
		_ = conf.SetCache(cacheKey, string(data), ytsCacheTTL)
	}

	utils.RespondSuccess(c, http.StatusOK, result)
}

func (h *MovieHandler) SearchFree(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.RespondError(c, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	cacheKey := fmt.Sprintf("free:search:%s:page:%d", query, page)
	if cached, err := conf.GetCache(cacheKey); err == nil {
		var result models.SearchResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			utils.RespondSuccess(c, http.StatusOK, result)
			return
		}
	}

	type searchResult struct {
		result *models.SearchResult
		err    error
	}

	legitCh := make(chan searchResult, 1)
	archiveCh := make(chan searchResult, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r, err := h.legit.Search(query, page)
		legitCh <- searchResult{r, err}
	}()

	go func() {
		defer wg.Done()
		r, err := h.archive.Search(query, page)
		archiveCh <- searchResult{r, err}
	}()

	wg.Wait()
	close(legitCh)
	close(archiveCh)

	legitRes := <-legitCh
	archiveRes := <-archiveCh

	seen := make(map[string]struct{})
	var merged []models.Movie

	addMovies := func(movies []models.Movie) {
		for _, m := range movies {
			key := normalizeTitle(m.Title)
			if key == "" {
				continue
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, m)
		}
	}

	if legitRes.err == nil && legitRes.result != nil {
		addMovies(legitRes.result.Results)
	}
	if archiveRes.err == nil && archiveRes.result != nil {
		addMovies(archiveRes.result.Results)
	}

	if legitRes.err != nil && archiveRes.err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch free torrent results")
		return
	}

	totalPages := 1
	if archiveRes.result != nil && archiveRes.result.TotalPages > totalPages {
		totalPages = archiveRes.result.TotalPages
	}

	result := models.SearchResult{
		Page:       page,
		TotalPages: totalPages,
		Results:    merged,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = conf.SetCache(cacheKey, string(data), freeCacheTTL)
	}

	utils.RespondSuccess(c, http.StatusOK, result)
}

func normalizeTitle(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "library-service"})
}
