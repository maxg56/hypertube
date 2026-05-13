package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"library-service/src/client"
	"library-service/src/conf"
	"library-service/src/models"
	"library-service/src/utils"
)

func (h *MovieHandler) fetchMovies(params client.ListParams, year int) (models.CursorResult, error) {
	searchResult, err := h.yts.List(params)
	if err != nil {
		return models.CursorResult{}, err
	}
	result := models.CursorResult{
		Results: searchResult.Results,
		Total:   searchResult.TotalCount,
	}
	if params.Page < searchResult.TotalPages {
		nextPage := params.Page + 1
		if year > 0 {
			nextPage = params.Page + 10
		}
		result.NextCursor = encodeCursor(nextPage)
	}
	return result, nil
}

func (h *MovieHandler) Movies(c *gin.Context) {
	query := c.Query("q")
	genre := c.Query("genre")
	ratingStr := c.Query("rating")
	yearStr := c.Query("year")
	sortBy := c.DefaultQuery("sort_by", "seeds")
	cursor := c.Query("cursor")
	userIDStr := c.GetHeader("X-User-ID")

	page := decodeCursor(cursor)

	var minRating float64
	if ratingStr != "" {
		minRating, _ = strconv.ParseFloat(ratingStr, 64)
	}
	var year int
	if yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}

	cacheKey := fmt.Sprintf("movies:q:%s:genre:%s:rating:%.1f:year:%d:sort:%s:page:%d",
		query, genre, minRating, year, sortBy, page)

	params := client.ListParams{
		Query:     query,
		Genre:     genre,
		MinRating: minRating,
		Year:      year,
		SortBy:    sortBy,
		Page:      page,
	}

	var result models.CursorResult
	if cacheGet(cacheKey, &result) {
		cacheRefreshIfStale(cacheKey, ytsCacheTTL, func() (interface{}, error) {
			return h.fetchMovies(params, year)
		})
		applyWatchedFlags(result.Results, userIDStr)
		utils.RespondSuccess(c, http.StatusOK, result)
		return
	}

	result, err := h.fetchMovies(params, year)
	if err != nil {
		utils.RespondError(c, http.StatusBadGateway, "failed to fetch movies")
		return
	}

	cacheSet(cacheKey, result, ytsCacheTTL)
	applyWatchedFlags(result.Results, userIDStr)
	utils.RespondSuccess(c, http.StatusOK, result)
}

func applyWatchedFlags(results []models.Movie, userIDStr string) {
	if conf.DB == nil || userIDStr == "" || len(results) == 0 {
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		return
	}

	tmdbIDs := make([]int, 0, len(results))
	for _, movie := range results {
		if movie.ID > 0 {
			tmdbIDs = append(tmdbIDs, movie.ID)
		}
	}
	if len(tmdbIDs) == 0 {
		return
	}

	var watchedIDs []int
	if err := conf.DB.Raw(`
		SELECT DISTINCT m.tmdb_id
		FROM watch_history wh
		INNER JOIN movies m ON m.id = wh.movie_id
		WHERE wh.user_id = ? AND m.tmdb_id IN ?
	`, userID, tmdbIDs).Scan(&watchedIDs).Error; err != nil {
		return
	}

	watchedSet := make(map[int]struct{}, len(watchedIDs))
	for _, id := range watchedIDs {
		watchedSet[id] = struct{}{}
	}

	for i := range results {
		_, results[i].Watched = watchedSet[results[i].ID]
	}
}

func encodeCursor(page int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(page)))
}

func decodeCursor(s string) int {
	if s == "" {
		return 1
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 1
	}
	page, _ := strconv.Atoi(string(b))
	if page < 1 {
		return 1
	}
	return page
}
