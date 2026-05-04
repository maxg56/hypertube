package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"library-service/src/client"
	"library-service/src/conf"
	"library-service/src/utils"
)

const (
	ytsCacheTTL  = 1 * time.Hour
	freeCacheTTL = 2 * time.Hour
)

type MovieHandler struct {
	tmdb    *client.TMDbClient
	omdb    *client.OMDbClient
	yts     *client.YTSClient
	legit   *client.LegitTorrentsClient
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
