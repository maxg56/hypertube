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

type favoriteMovieRow struct {
	TmdbID      int     `json:"tmdb_id"     gorm:"column:tmdb_id"`
	Title       string  `json:"title"       gorm:"column:title"`
	PosterURL   string  `json:"poster_url"  gorm:"column:poster_url"`
	Rating      float64 `json:"rating"      gorm:"column:rating"`
	Language    string  `json:"language"    gorm:"column:language"`
	ReleaseDate string  `json:"release_date" gorm:"column:release_date"`
}

func listFavoritesForUser(c *gin.Context, userID uint) {
	pagination := utils.ParsePaginationParams(c)

	var total int64
	if err := conf.DB.Model(&models.Favorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to count favorites")
		return
	}

	var rows []favoriteMovieRow
	err := conf.DB.Raw(`
		SELECT
			f.tmdb_id,
			COALESCE(m.title, '')                               AS title,
			COALESCE(m.poster_path, '')                         AS poster_url,
			COALESCE(m.rating, 0)                               AS rating,
			COALESCE(m.language, '')                            AS language,
			COALESCE(to_char(m.release_date, 'YYYY-MM-DD'), '') AS release_date
		FROM favorites f
		LEFT JOIN movies m ON m.tmdb_id = f.tmdb_id
		WHERE f.user_id = ?
		ORDER BY f.added_at DESC
		LIMIT ? OFFSET ?
	`, userID, pagination.Limit, pagination.Offset).Scan(&rows).Error
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch favorites")
		return
	}

	if rows == nil {
		rows = []favoriteMovieRow{}
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"favorites":  rows,
		"pagination": utils.NewPagination(total, pagination.Limit, pagination.Offset),
	})
}

// ListFavoritesHandler handles GET /api/v1/users/favorites (own favorites)
func ListFavoritesHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}
	listFavoritesForUser(c, userID)
}

// ListUserFavoritesHandler handles GET /api/v1/users/:id/favorites (any user's public favorites)
func ListUserFavoritesHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var user models.User
	if err := conf.DB.Select("is_public, favorites_public").First(&user, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}
	if !user.IsPublic || !user.FavoritesPublic {
		utils.RespondError(c, http.StatusForbidden, "favorites are private")
		return
	}

	listFavoritesForUser(c, uint(id))
}

// AddFavoriteHandler handles POST /api/v1/users/favorites
func AddFavoriteHandler(c *gin.Context) {
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

	fav := models.Favorite{UserID: int(userID), TmdbID: body.TmdbID}
	result := conf.DB.Where(models.Favorite{UserID: int(userID), TmdbID: body.TmdbID}).FirstOrCreate(&fav)
	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to add favorite")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, gin.H{"tmdb_id": body.TmdbID})
}

// RemoveFavoriteHandler handles DELETE /api/v1/users/favorites/:tmdbId
func RemoveFavoriteHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	tmdbID, convErr := strconv.Atoi(c.Param("tmdbId"))
	if convErr != nil || tmdbID == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	if err := conf.DB.Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).Delete(&models.Favorite{}).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to remove favorite")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"tmdb_id": tmdbID})
}

// CheckFavoriteHandler handles GET /api/v1/users/favorites/:tmdbId
func CheckFavoriteHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	tmdbID, convErr := strconv.Atoi(c.Param("tmdbId"))
	if convErr != nil || tmdbID == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	var fav models.Favorite
	err = conf.DB.Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).First(&fav).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"is_favorite": !errors.Is(err, gorm.ErrRecordNotFound)})
}
