package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

type watchLaterMovieRow struct {
	TmdbID      int     `json:"tmdb_id"      gorm:"column:tmdb_id"`
	Title       string  `json:"title"        gorm:"column:title"`
	PosterURL   string  `json:"poster_url"   gorm:"column:poster_url"`
	Rating      float64 `json:"rating"       gorm:"column:rating"`
	Language    string  `json:"language"     gorm:"column:language"`
	ReleaseDate string  `json:"release_date" gorm:"column:release_date"`
}

// ListWatchLaterHandler handles GET /api/v1/users/watch-later
func ListWatchLaterHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	pagination := utils.ParsePaginationParams(c)

	var total int64
	if err := conf.DB.Model(&models.WatchLater{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to count watch later")
		return
	}

	var rows []watchLaterMovieRow
	err = conf.DB.Raw(`
		SELECT
			wl.tmdb_id,
			COALESCE(m.title, '')                               AS title,
			COALESCE(m.poster_path, '')                         AS poster_url,
			COALESCE(m.rating, 0)                               AS rating,
			COALESCE(m.language, '')                            AS language,
			COALESCE(to_char(m.release_date, 'YYYY-MM-DD'), '') AS release_date
		FROM watch_later wl
		LEFT JOIN movies m ON m.tmdb_id = wl.tmdb_id
		WHERE wl.user_id = ?
		ORDER BY wl.added_at DESC
		LIMIT ? OFFSET ?
	`, userID, pagination.Limit, pagination.Offset).Scan(&rows).Error
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch watch later")
		return
	}

	if rows == nil {
		rows = []watchLaterMovieRow{}
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"items":      rows,
		"pagination": utils.NewPagination(total, pagination.Limit, pagination.Offset),
	})
}

// AddWatchLaterHandler handles POST /api/v1/users/watch-later
func AddWatchLaterHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	var body struct {
		TmdbID int `json:"tmdb_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.TmdbID == 0 {
		utils.RespondError(c, http.StatusBadRequest, "tmdb_id is required")
		return
	}

	item := models.WatchLater{UserID: int(userID), TmdbID: body.TmdbID}
	result := conf.DB.Where(models.WatchLater{UserID: int(userID), TmdbID: body.TmdbID}).FirstOrCreate(&item)
	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to add to watch later")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, gin.H{"tmdb_id": body.TmdbID})
}

// RemoveWatchLaterHandler handles DELETE /api/v1/users/watch-later/:tmdbId
func RemoveWatchLaterHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	tmdbID, convErr := strconv.Atoi(c.Param("tmdbId"))
	if convErr != nil || tmdbID == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	if err := conf.DB.Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).Delete(&models.WatchLater{}).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to remove from watch later")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"tmdb_id": tmdbID})
}

// CheckWatchLaterHandler handles GET /api/v1/users/watch-later/:tmdbId
func CheckWatchLaterHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	tmdbID, convErr := strconv.Atoi(c.Param("tmdbId"))
	if convErr != nil || tmdbID == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	var item models.WatchLater
	err = conf.DB.Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).First(&item).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"in_watch_later": !errors.Is(err, gorm.ErrRecordNotFound)})
}
