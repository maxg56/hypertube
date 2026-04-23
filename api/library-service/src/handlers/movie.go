package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"library-service/src/client"
	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

type MovieHandler struct {
	tmdb *client.TMDbClient
	omdb *client.OMDbClient
}

func NewMovieHandler() *MovieHandler {
	return &MovieHandler{
		tmdb: client.NewTMDbClient(),
		omdb: client.NewOMDbClient(),
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

	if movie == nil && h.omdb.Available() {
		// TMDb unavailable or movie not found — try OMDb by title if we have one
		// For a pure ID-based fallback we skip (OMDb uses imdb IDs, not TMDb IDs)
	}

	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch movie details")
		return
	}
	if movie == nil {
		utils.RespondError(c, http.StatusNotFound, "movie not found")
		return
	}

	if data, err := json.Marshal(movie); err == nil {
		_ = conf.SetCache(cacheKey, string(data), conf.MovieCacheTTL)
	}

	utils.RespondSuccess(c, http.StatusOK, movie)
}

func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "library-service"})
}
