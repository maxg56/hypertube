package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

func (h *MovieHandler) GetMovie(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	cacheKey := fmt.Sprintf("movie:%d", id)
	var movie models.Movie
	if cacheGet(cacheKey, &movie) {
		utils.RespondSuccess(c, http.StatusOK, movie)
		return
	}

	if !h.tmdb.Available() {
		utils.RespondError(c, http.StatusServiceUnavailable, "no metadata provider available")
		return
	}

	result, err := h.tmdb.GetMovie(id)
	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch movie details")
		return
	}
	if result == nil {
		utils.RespondError(c, http.StatusNotFound, "movie not found")
		return
	}

	if result.IMDbID != "" {
		if ytsMovie, ytsErr := h.yts.GetMovieByIMDbID(result.IMDbID); ytsErr == nil && ytsMovie != nil {
			result.Torrents = ytsMovie.Torrents
		}
	}

	cacheSet(cacheKey, result, conf.MovieCacheTTL)
	utils.RespondSuccess(c, http.StatusOK, result)
}
