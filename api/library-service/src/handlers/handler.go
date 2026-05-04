package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"library-service/src/client"
	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

const (
	ytsCacheTTL  = 1 * time.Hour
	freeCacheTTL = 2 * time.Hour
)

type tmdbClient interface {
	Available() bool
	Search(query string, page int) (*models.SearchResult, error)
	GetMovie(id int) (*models.Movie, error)
}

type searchClient interface {
	Available() bool
	Search(query string, page int) (*models.SearchResult, error)
}

type ytsClient interface {
	Search(query string, page int) (*models.SearchResult, error)
	GetMovieByIMDbID(imdbID string) (*models.Movie, error)
}

type freeClient interface {
	Search(query string, page int) (*models.SearchResult, error)
}

type MovieHandler struct {
	tmdb    tmdbClient
	omdb    searchClient
	yts     ytsClient
	legit   freeClient
	archive freeClient
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

func parseQueryPage(c *gin.Context) (query string, page int, ok bool) {
	query = c.Query("q")
	if query == "" {
		utils.RespondError(c, http.StatusBadRequest, "query parameter 'q' is required")
		return "", 0, false
	}
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	return query, page, true
}

func cacheGet(key string, dst interface{}) bool {
	cached, err := conf.GetCache(key)
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(cached), dst) == nil
}

func cacheSet(key string, v interface{}, ttl time.Duration) {
	if data, err := json.Marshal(v); err == nil {
		_ = conf.SetCache(key, string(data), ttl)
	}
}
